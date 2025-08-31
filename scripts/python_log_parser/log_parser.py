#!/usr/bin/env python3
"""
Agent Communication Flow Visualizer - Enhanced Version
"""

import json
import re
import sys
from datetime import datetime
from collections import defaultdict
import subprocess

try:
    import matplotlib.pyplot as plt
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
        self.from_agent = from_parts[0] if len(from_parts) > 0 and from_parts[0] else ''
        self.from_type = from_parts[1] if len(from_parts) > 1 and from_parts[1] else ''

        to_parts = to_str.split('/')
        self.to_agent = to_parts[0] if len(to_parts) > 0 and to_parts[0] else ''
        self.to_type = to_parts[1] if len(to_parts) > 1 and to_parts[1] else ''

        # Try to extract from headers in payload if not found
        if not self.from_type and self.payload_preview:
            try:
                payload = json.loads(self.payload_preview)
                if 'from_agent_type' in payload:
                    self.from_type = payload['from_agent_type']
                if 'to_agent_type' in payload:
                    self.to_type = payload['to_agent_type']
            except:
                pass

class LogParser:
    def __init__(self):
        self.messages = []
        self.agents = {}  # agent_id -> agent_info
        self.orchestrations = defaultdict(list)
        self.awaited_steps = defaultdict(list)
        self.spawned_agents = {}  # Track spawned agents

    def parse_k8s_logs(self, namespace="ai-persona-system", label="app=agent-chassis"):
        """Fetch and parse Kubernetes logs"""
        cmd = f"kubectl -n {namespace} logs -l {label} --prefix=true --tail=3000"

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

            # Track spawned agents
            if 'Successfully spawned agent' in log_entry.get('msg', ''):
                agent_id = log_entry.get('DEBUG_SPAWN_31: agent_id', '')
                agent_type = log_entry.get('DEBUG_SPAWN_31: agent_type', '')
                if agent_id and agent_type:
                    self.agents[agent_id] = {
                        'type': agent_type,
                        'id': agent_id[:8],
                        'full_id': agent_id
                    }
                    print(f"   🤖 Spawned: {agent_type} [{agent_id[:8]}]")

            elif 'All agents spawned' in log_entry.get('msg', ''):
                spawned = log_entry.get('DEBUG_SPAWN_14: spawned_agents', {})
                for role, agent_id in spawned.items():
                    if agent_id:
                        self.spawned_agents[agent_id] = role

            # Look for MESSAGE_TRACE entries
            if log_entry.get('msg') == 'MESSAGE_TRACE':
                msg = AgentMessage(timestamp, log_entry)

                # Try to identify agents from context
                if not msg.from_type and msg.from_agent in self.agents:
                    msg.from_type = self.agents[msg.from_agent]['type']
                if not msg.to_type and msg.to_agent in self.agents:
                    msg.to_type = self.agents[msg.to_agent]['type']

                # Default to "orchestrator" for the main generic agent
                if msg.topic and 'generic' in msg.topic and not msg.from_type:
                    msg.from_type = 'orchestrator'
                if msg.topic and 'generic' in msg.topic and not msg.to_type:
                    msg.to_type = 'orchestrator'

                self.messages.append(msg)

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

        except Exception as e:
            pass

    def format_agent_name(self, agent_id, agent_type):
        """Format agent name for display"""
        if agent_type:
            if agent_id:
                return f"{agent_type}[{agent_id[:8] if len(agent_id) > 8 else agent_id}]"
            return agent_type
        elif agent_id:
            # Try to look up type from our tracked agents
            if agent_id in self.agents:
                return f"{self.agents[agent_id]['type']}[{agent_id[:8]}]"
            elif agent_id in self.spawned_agents:
                return f"{self.spawned_agents[agent_id]}[{agent_id[:8]}]"
            return f"agent[{agent_id[:8]}]"
        return "SYSTEM"

    def format_agent_display(self, agent_id, agent_type):
        """Format agent for display with ID"""
        if agent_id and len(agent_id) > 8:
            short_id = agent_id[:8]
        else:
            short_id = agent_id if agent_id else "unknown"

        if agent_type:
            return f"{agent_type}[{short_id}]"
        elif agent_id in self.agents:
            info = self.agents[agent_id]
            return f"{info['type']}[{short_id}]"
        elif agent_id in self.spawned_agents:
            role = self.spawned_agents[agent_id]
            return f"{role}[{short_id}]"
        elif agent_id:
            return f"agent[{short_id}]"
        else:
            return "SYSTEM"

    def print_message(self, msg):
        """Print a single message with formatting"""
        # Determine arrow and color based on direction
        arrows = {
            'sending_response': ('→', '\033[92m'),  # Green
            'processing_response': ('←', '\033[94m'),  # Blue
            'received': ('↓', '\033[93m'),  # Yellow
            'sending': ('↑', '\033[96m'),  # Cyan
        }
        arrow, color = arrows.get(msg.direction, ('↔', '\033[90m'))

        # from_str = self.format_agent_name(msg.from_agent, msg.from_type)
        # to_str = self.format_agent_name(msg.to_agent, msg.to_type)

        from_str = self.format_agent_display(msg.from_agent, msg.from_type)
        to_str = self.format_agent_display(msg.to_agent, msg.to_type)

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
                        break
                    elif await_info['awaited']:
                        print(f"   ❌ MISMATCH: Expected {await_info['awaited'][0][:8]}...")
                        break

        if msg.topic:
            print(f"   Topic: {msg.topic}")

        if msg.payload_preview:
            try:
                payload = json.loads(msg.payload_preview)
                if 'action' in payload:
                    print(f"   Action: {payload['action']}")
                if 'data' in payload and isinstance(payload['data'], dict):
                    print(f"   Data keys: {list(payload['data'].keys())}")
            except:
                print(f"   Payload: {msg.payload_preview[:100]}...")

        print('\033[0m')  # Reset color

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

        # Show spawned agents
        if self.agents:
            print("🤖 Active Agents:")
            for agent_id, info in self.agents.items():
                role = self.spawned_agents.get(agent_id, 'unknown')
                print(f"   - {info['type']} ({role}) [{info['id']}]")
            print()

        # Group messages by orchestration
        for orch_id, orch_messages in self.orchestrations.items():
            if correlation_id and not any(m.correlation_id == correlation_id for m in orch_messages):
                continue

            print(f"\n📋 Orchestration: {orch_id[:16]}...")

            # Show awaited steps for this orchestration
            if orch_id in self.awaited_steps:
                print(f"   ⏳ Awaited Steps:")
                for await_info in self.awaited_steps[orch_id]:
                    print(f"      - Action: {await_info['action']}")
                    for req_id in await_info['awaited']:
                        print(f"        Request ID: {req_id[:16]}...")

            print(f"\n   Messages:")
            for msg in sorted(orch_messages, key=lambda x: x.timestamp):
                self.print_message(msg)

    def print_summary(self):
        """Print summary with request ID analysis"""
        print(f"\n{'='*100}")
        print("Summary Analysis")
        print(f"{'='*100}\n")

        print(f"Total Messages: {len(self.messages)}")
        print(f"Total Agents Tracked: {len(self.agents)}")
        print(f"Total Orchestrations: {len(self.orchestrations)}")

        # Analyze request/response matching
        print("\n🔍 Request/Response Analysis:")

        for orch_id, await_list in self.awaited_steps.items():
            print(f"\n   Orchestration {orch_id[:16]}...")
            for await_info in await_list:
                if await_info['awaited']:
                    for awaited_id in await_info['awaited']:
                        # Check if we have a response for this
                        has_response = any(m.in_response_to == awaited_id for m in self.messages)
                        status = "✅ Received" if has_response else "❌ Missing"
                        print(f"      Awaiting {awaited_id[:16]}... from {await_info['action']}: {status}")

def main():
    parser = LogParser()

    print("Fetching logs from Kubernetes...")
    parser.parse_k8s_logs()

    if not parser.messages:
        print("No MESSAGE_TRACE entries found.")
        return

    parser.print_summary()
    parser.print_flow()

if __name__ == "__main__":
    main()