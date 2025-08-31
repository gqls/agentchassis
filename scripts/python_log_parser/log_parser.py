#!/usr/bin/env python3
"""
Agent Communication Flow Visualizer - Fixed Version
"""

import json
import re
import sys
from datetime import datetime
from collections import defaultdict
import subprocess

try:
    import matplotlib.pyplot as plt
    import matplotlib.patches as mpatches
    from matplotlib.patches import FancyBboxPatch, FancyArrowPatch
    import networkx as nx
    HAVE_MATPLOTLIB = True
except ImportError:
    HAVE_MATPLOTLIB = False

class AgentMessage:
    def __init__(self, timestamp, log_entry):
        self.timestamp = timestamp
        self.raw = log_entry

        # Extract from log entry
        self.correlation_id = log_entry.get('correlation_id', '')
        self.orchestration_id = log_entry.get('orchestration_id', '')
        self.request_id = log_entry.get('request_id', '')
        self.in_response_to = log_entry.get('in_response_to', '')
        self.direction = log_entry.get('direction', '')
        self.topic = log_entry.get('topic', '')
        self.message_type = log_entry.get('message_type', '')
        self.action = log_entry.get('action', '')
        self.payload_preview = log_entry.get('payload_preview', '')

        # Parse from/to from the log format "agent_id/agent_type"
        from_str = log_entry.get('from', '/')
        to_str = log_entry.get('to', '/')

        from_parts = from_str.split('/')
        self.from_agent = from_parts[0] if len(from_parts) > 0 else ''
        self.from_type = from_parts[1] if len(from_parts) > 1 else ''

        to_parts = to_str.split('/')
        self.to_agent = to_parts[0] if len(to_parts) > 0 else ''
        self.to_type = to_parts[1] if len(to_parts) > 1 else ''

class LogParser:
    def __init__(self):
        self.messages = []
        self.agents = set()
        self.orchestrations = defaultdict(list)
        self.awaited_steps = defaultdict(list)  # Track what each orchestration is waiting for

    def parse_k8s_logs(self, namespace="ai-persona-system", label="app=agent-chassis"):
        """Fetch and parse Kubernetes logs"""
        cmd = f"kubectl -n {namespace} logs -l {label} --prefix=true --tail=2000"

        try:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
            if result.returncode != 0:
                print(f"Error fetching logs: {result.stderr}")
                return

            for line in result.stdout.split('\n'):
                self.parse_line(line)

        except Exception as e:
            print(f"Error: {e}")

    def parse_line(self, line):
        """Parse a single log line"""
        try:
            # Extract JSON part
            json_match = re.search(r'\{.*\}$', line)
            if not json_match:
                return

            log_entry = json.loads(json_match.group())

            # Extract timestamp
            timestamp = log_entry.get('ts', '')

            # Look for different types of log entries
            if log_entry.get('msg') == 'MESSAGE_TRACE':
                msg = AgentMessage(timestamp, log_entry)
                self.messages.append(msg)

                # Track agents
                if msg.from_agent:
                    self.agents.add((msg.from_agent, msg.from_type))
                if msg.to_agent:
                    self.agents.add((msg.to_agent, msg.to_type))

                # Group by orchestration
                if msg.orchestration_id:
                    self.orchestrations[msg.orchestration_id].append(msg)

            elif log_entry.get('msg') == 'AWAITED_STEPS_CHANGED':
                # Track what steps are being awaited
                orch_id = log_entry.get('orchestration_id', '')
                awaited = log_entry.get('awaited_steps', [])
                action = log_entry.get('for_action', '')
                if orch_id:
                    self.awaited_steps[orch_id].append({
                        'timestamp': timestamp,
                        'awaited': awaited,
                        'action': action,
                        'request_id': log_entry.get('request_id', '')
                    })

            elif 'spawn_group' in str(log_entry.get('msg', '')).lower():
                # Capture spawn group details
                if 'request_id' in log_entry:
                    print(f"Found spawn_group with request_id: {log_entry.get('request_id')}")

        except Exception as e:
            # Skip malformed lines
            pass

    def print_flow(self, correlation_id=None):
        """Print message flow with request ID tracking"""
        messages = self.messages
        if correlation_id:
            messages = [m for m in messages if m.correlation_id == correlation_id]

        print(f"\n{'='*100}")
        print(f"Agent Communication Flow Analysis")
        if correlation_id:
            print(f"Correlation ID: {correlation_id}")
        print(f"{'='*100}\n")

        # Group messages by orchestration
        for orch_id, orch_messages in self.orchestrations.items():
            if correlation_id and not any(m.correlation_id == correlation_id for m in orch_messages):
                continue

            print(f"\n📋 Orchestration: {orch_id}")

            # Show awaited steps for this orchestration
            if orch_id in self.awaited_steps:
                print(f"   ⏳ Awaited Steps:")
                for await_info in self.awaited_steps[orch_id]:
                    print(f"      - Action: {await_info['action']}")
                    print(f"        Request IDs: {await_info['awaited']}")
                    print(f"        Set at: {await_info['timestamp']}")

            print(f"\n   Messages:")
            for msg in sorted(orch_messages, key=lambda x: x.timestamp):
                self.print_message(msg)

    def print_message(self, msg):
        """Print a single message with formatting"""
        # Determine arrow and color based on direction
        if msg.direction == 'sending_response':
            arrow = "→"
            color = '\033[92m'  # Green
        elif msg.direction == 'processing_response':
            arrow = "←"
            color = '\033[94m'  # Blue
        elif msg.direction == 'received':
            arrow = "↓"
            color = '\033[93m'  # Yellow
        else:
            arrow = "↔"
            color = '\033[90m'  # Gray

        from_str = f"{msg.from_type}[{msg.from_agent[:8]}]" if msg.from_agent else "SYSTEM"
        to_str = f"{msg.to_type}[{msg.to_agent[:8]}]" if msg.to_agent else "SYSTEM"

        print(f"\n{color}   {msg.timestamp}")
        print(f"   {from_str} {arrow} {to_str}")
        print(f"   Direction: {msg.direction}")

        if msg.request_id:
            print(f"   Request ID: {msg.request_id}")
        if msg.in_response_to:
            print(f"   In Response To: {msg.in_response_to}")

            # Check if this matches any awaited steps
            orch_id = msg.orchestration_id
            if orch_id in self.awaited_steps:
                for await_info in self.awaited_steps[orch_id]:
                    if msg.in_response_to in await_info['awaited']:
                        print(f"   ✅ MATCHES awaited step from {await_info['action']}")
                    elif await_info['awaited']:
                        print(f"   ❌ MISMATCH: Expected {await_info['awaited'][0]}")

        if msg.topic:
            print(f"   Topic: {msg.topic}")

        if msg.payload_preview:
            # Parse and show key parts of payload
            try:
                payload = json.loads(msg.payload_preview)
                if 'action' in payload:
                    print(f"   Action: {payload['action']}")
                if 'data' in payload:
                    print(f"   Data keys: {list(payload['data'].keys())}")
            except:
                print(f"   Payload: {msg.payload_preview[:100]}...")

        print('\033[0m')  # Reset color

    def print_summary(self):
        """Print summary with request ID analysis"""
        print(f"\n{'='*100}")
        print("Summary Analysis")
        print(f"{'='*100}\n")

        print(f"Total Messages: {len(self.messages)}")
        print(f"Total Agents: {len(self.agents)}")
        print(f"Total Orchestrations: {len(self.orchestrations)}")

        # Analyze request/response matching
        print("\n🔍 Request/Response Analysis:")
        request_ids = set()
        response_ids = set()

        for msg in self.messages:
            if msg.request_id:
                request_ids.add(msg.request_id)
            if msg.in_response_to:
                response_ids.add(msg.in_response_to)

        print(f"   Unique Request IDs: {len(request_ids)}")
        print(f"   Unique Response IDs: {len(response_ids)}")

        # Find mismatches
        unmatched_requests = request_ids - response_ids
        unmatched_responses = response_ids - request_ids

        if unmatched_requests:
            print(f"\n   ⚠️  Requests without responses:")
            for req_id in unmatched_requests:
                print(f"      - {req_id}")

        if unmatched_responses:
            print(f"\n   ⚠️  Responses without matching requests:")
            for resp_id in unmatched_responses:
                print(f"      - {resp_id}")

        # Check awaited vs received
        print("\n📊 Awaited Steps Analysis:")
        for orch_id, await_list in self.awaited_steps.items():
            print(f"   Orchestration {orch_id[:8]}...")
            for await_info in await_list:
                if await_info['awaited']:
                    awaited_id = await_info['awaited'][0]
                    # Check if we have a response for this
                    has_response = any(m.in_response_to == awaited_id for m in self.messages)
                    status = "✅ Received" if has_response else "❌ Missing"
                    print(f"      Awaiting {awaited_id[:8]}... from {await_info['action']}: {status}")

def main():
    parser = LogParser()

    print("Fetching logs from Kubernetes...")
    parser.parse_k8s_logs()

    if not parser.messages:
        print("No MESSAGE_TRACE entries found in logs.")
        print("\nTip: Make sure MESSAGE_TRACE logging is enabled in your application.")
        return

    # Print summary
    parser.print_summary()

    # Print detailed flow
    parser.print_flow()

    # Offer to filter by correlation ID
    if parser.orchestrations:
        print(f"\n{'='*100}")
        print("Available orchestrations for detailed analysis:")
        for i, (orch_id, msgs) in enumerate(list(parser.orchestrations.items())[:10]):
            print(f"  {i+1}. {orch_id} ({len(msgs)} messages)")

        choice = input("\nEnter orchestration number for details (or press Enter to skip): ").strip()
        if choice.isdigit():
            idx = int(choice) - 1
            if 0 <= idx < len(parser.orchestrations):
                orch_id = list(parser.orchestrations.keys())[idx]
                print(f"\nDetailed view for orchestration {orch_id}:")
                parser.print_flow(correlation_id=None)  # Show all for now

if __name__ == "__main__":
    main()