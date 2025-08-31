#!/usr/bin/env python3
"""
Agent Communication Flow Visualizer - Enhanced Version with Detailed Tracing
"""

import json
import re
import sys
from datetime import datetime
from collections import defaultdict, OrderedDict
import subprocess

class AgentMessage:
    def __init__(self, timestamp, log_entry, source_func=None):
        self.timestamp = timestamp
        self.raw = log_entry
        self.source_func = source_func  # Track where this log came from

        # Extract all IDs
        self.correlation_id = log_entry.get('correlation_id', '')
        self.orchestration_id = log_entry.get('orchestration_id', '')
        self.request_id = log_entry.get('request_id', '')
        self.in_response_to = log_entry.get('in_response_to', '')
        self.direction = log_entry.get('direction', '')
        self.topic = log_entry.get('topic', '')
        self.message_type = log_entry.get('message_type', '')
        self.action = log_entry.get('action', '')
        self.payload_preview = log_entry.get('payload_preview', '')

        # Extract agent IDs - look in multiple places
        self.from_agent_id = log_entry.get('from_agent_id', '')
        self.from_agent_type = log_entry.get('from_agent_type', '')
        self.to_agent_id = log_entry.get('to_agent_id', '')
        self.to_agent_type = log_entry.get('to_agent_type', '')

        # Parse from/to strings if present
        from_str = log_entry.get('from', '/')
        to_str = log_entry.get('to', '/')

        if '/' in from_str:
            parts = from_str.split('/')
            if not self.from_agent_id and len(parts) > 0:
                self.from_agent_id = parts[0]
            if not self.from_agent_type and len(parts) > 1:
                self.from_agent_type = parts[1]

        if '/' in to_str:
            parts = to_str.split('/')
            if not self.to_agent_id and len(parts) > 0:
                self.to_agent_id = parts[0]
            if not self.to_agent_type and len(parts) > 1:
                self.to_agent_type = parts[1]

class FlowStep:
    """Represents a step in the execution flow"""
    def __init__(self, timestamp, step_type, details, source_func):
        self.timestamp = timestamp
        self.step_type = step_type  # e.g., "PRE_GENERATE_REQUEST", "AWAIT_SETUP", etc.
        self.details = details
        self.source_func = source_func

class LogParser:
    def __init__(self):
        self.messages = []
        self.agents = {}  # agent_id -> agent_info
        self.orchestrations = defaultdict(list)
        self.awaited_steps = defaultdict(list)
        self.spawned_agents = {}
        self.flow_steps = []  # Track execution flow
        self.request_mappings = {}  # Map original request to pre-generated ones

    def parse_k8s_logs(self, namespace="ai-persona-system", label="app=agent-chassis"):
        """Fetch and parse Kubernetes logs"""
        cmd = f"kubectl -n {namespace} logs -l {label} --prefix=true --tail=5000"

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

            # Extract timestamp and caller (source function)
            timestamp = log_entry.get('ts', '')
            caller = log_entry.get('caller', '')
            source_func = caller.split('/')[-1] if caller else 'unknown'

            # Track different types of log entries
            msg = log_entry.get('msg', '')

            # Track spawned agents with their IDs
            if 'Successfully spawned agent' in msg:
                agent_id = log_entry.get('DEBUG_SPAWN_12: agent_id') or \
                           log_entry.get('DEBUG_SPAWN_31: agent_id', '')
                agent_type = log_entry.get('DEBUG_SPAWN_12: agent_type') or \
                             log_entry.get('DEBUG_SPAWN_31: agent_type', '')
                role = log_entry.get('DEBUG_SPAWN_12: role', '')

                if agent_id and agent_type:
                    self.agents[agent_id] = {
                        'type': agent_type,
                        'id': agent_id[:8],
                        'full_id': agent_id,
                        'role': role
                    }
                    print(f"   🤖 Spawned: {agent_type} ({role}) [{agent_id[:8]}]")

            # Track pre-generated request IDs
            elif 'Pre-generated request ID' in msg:
                original = log_entry.get('request_id', '')
                generated = log_entry.get('action_request_id', '') or log_entry.get('request_id', '')
                if original and generated and original != generated:
                    self.request_mappings[original] = generated
                    self.flow_steps.append(FlowStep(
                        timestamp,
                        "PRE_GENERATE_REQUEST",
                        {
                            'original': original[:16],
                            'generated': generated[:16],
                            'action': log_entry.get('action', '')
                        },
                        source_func
                    ))

            # Track when awaited steps are set
            elif msg == 'AWAITED_STEPS_CHANGED':
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
                    self.flow_steps.append(FlowStep(
                        timestamp,
                        "AWAIT_SETUP",
                        {
                            'orchestration_id': orch_id[:16],
                            'awaited_requests': [r[:16] for r in awaited],
                            'action': action
                        },
                        source_func
                    ))

            # Track MESSAGE_TRACE entries
            elif msg == 'MESSAGE_TRACE':
                # Extract agent IDs from headers if present
                from_agent_id = log_entry.get('from_agent_id', '')
                to_agent_id = log_entry.get('to_agent_id', '')

                msg_obj = AgentMessage(timestamp, log_entry, source_func)

                # Override with explicit IDs if available
                if from_agent_id:
                    msg_obj.from_agent_id = from_agent_id
                if to_agent_id:
                    msg_obj.to_agent_id = to_agent_id

                self.messages.append(msg_obj)

                if msg_obj.orchestration_id:
                    self.orchestrations[msg_obj.orchestration_id].append(msg_obj)

            # Track response sending
            elif 'TRACE: Sending workflow response' in msg:
                self.flow_steps.append(FlowStep(
                    timestamp,
                    "SEND_RESPONSE",
                    {
                        'orchestration_id': log_entry.get('response_orch_id', '')[:16],
                        'in_response_to': log_entry.get('in_response_to', '')[:16] if log_entry.get('in_response_to') else 'none'
                    },
                    source_func
                ))

        except Exception as e:
            pass

    def format_agent_display(self, agent_id, agent_type, role=None):
        """Format agent for display with type, role, and ID"""
        if agent_id in self.agents:
            info = self.agents[agent_id]
            role_str = f"({info.get('role', 'unknown')})" if info.get('role') else ""
            return f"{info['type']}{role_str}[{info['id']}]"
        elif agent_type:
            short_id = agent_id[:8] if agent_id and len(agent_id) > 8 else agent_id or "unknown"
            return f"{agent_type}[{short_id}]"
        elif agent_id:
            short_id = agent_id[:8] if len(agent_id) > 8 else agent_id
            return f"agent[{short_id}]"
        else:
            return "SYSTEM"

    def print_message(self, msg):
        """Print a single message with detailed formatting"""
        arrows = {
            'sending_response': ('→', '\033[92m'),  # Green
            'processing_response': ('←', '\033[94m'),  # Blue
            'received': ('↓', '\033[93m'),  # Yellow
            'sending': ('→', '\033[96m'),  # Cyan
            'awaited_update': ('⏳', '\033[95m'),  # Purple
        }
        arrow, color = arrows.get(msg.direction, ('↔', '\033[90m'))

        from_str = self.format_agent_display(msg.from_agent_id, msg.from_agent_type)
        to_str = self.format_agent_display(msg.to_agent_id, msg.to_agent_type)

        print(f"\n{color}   [{msg.source_func}] {msg.timestamp}")
        print(f"   {from_str} {arrow} {to_str}")
        print(f"   Direction: {msg.direction}")

        if msg.request_id:
            print(f"   Request ID: {msg.request_id[:16]}...")
            # Check if this was pre-generated
            for orig, gen in self.request_mappings.items():
                if gen == msg.request_id:
                    print(f"   📝 (Pre-generated from: {orig[:16]}...)")

        if msg.in_response_to:
            print(f"   In Response To: {msg.in_response_to[:16]}...")

            # Check against awaited steps
            orch_id = msg.orchestration_id
            if orch_id in self.awaited_steps:
                for await_info in self.awaited_steps[orch_id]:
                    if msg.in_response_to in await_info['awaited']:
                        print(f"   ✅ MATCHES awaited step from {await_info['action']}")
                        break
                    elif await_info['awaited']:
                        expected = await_info['awaited'][0]
                        print(f"   ❌ MISMATCH: Expected {expected[:16]}...")
                        # Show the mapping if we know it
                        for orig, gen in self.request_mappings.items():
                            if gen == expected:
                                print(f"      (which was pre-generated from {orig[:16]}...)")
                        break

        if msg.topic:
            print(f"   Topic: {msg.topic}")

        # Parse and show payload details
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

    def print_flow_trace(self):
        """Print the execution flow trace"""
        if not self.flow_steps:
            return

        print(f"\n{'='*100}")
        print("Execution Flow Trace")
        print(f"{'='*100}\n")

        for step in sorted(self.flow_steps, key=lambda x: x.timestamp):
            color = {
                'PRE_GENERATE_REQUEST': '\033[95m',  # Purple
                'AWAIT_SETUP': '\033[93m',  # Yellow
                'SEND_RESPONSE': '\033[92m',  # Green
            }.get(step.step_type, '\033[90m')

            print(f"{color}[{step.source_func}] {step.timestamp}")
            print(f"   {step.step_type}")

            for key, value in step.details.items():
                print(f"      {key}: {value}")
            print('\033[0m')

    def print_flow(self, correlation_id=None):
        """Print message flow with enhanced details"""
        # First print the execution trace
        self.print_flow_trace()

        # Then print the message flow
        messages = self.messages
        if correlation_id:
            messages = [m for m in messages if m.correlation_id == correlation_id]

        print(f"\n{'='*100}")
        print(f"Agent Communication Flow")
        print(f"{'='*100}\n")

        # Show active agents
        if self.agents:
            print("🤖 Active Agents:")
            for agent_id, info in self.agents.items():
                role = f"({info.get('role', 'unknown')})" if info.get('role') else ""
                print(f"   - {info['type']} {role} [{info['id']}]")
            print()

        # Show request mappings
        if self.request_mappings:
            print("📝 Request ID Mappings:")
            for orig, gen in self.request_mappings.items():
                print(f"   {orig[:16]}... → {gen[:16]}... (pre-generated)")
            print()

        # Group by orchestration
        for orch_id, orch_messages in self.orchestrations.items():
            if correlation_id and not any(m.correlation_id == correlation_id for m in orch_messages):
                continue

            print(f"📋 Orchestration: {orch_id[:16]}...")

            # Show awaited steps
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
        """Print enhanced summary"""
        print(f"\n{'='*100}")
        print("Summary Analysis")
        print(f"{'='*100}\n")

        print(f"Total Messages: {len(self.messages)}")
        print(f"Total Agents Tracked: {len(self.agents)}")
        print(f"Total Orchestrations: {len(self.orchestrations)}")
        print(f"Request ID Mappings: {len(self.request_mappings)}")

        # Analyze request/response matching
        print("\n🔍 Request/Response Analysis:")

        for orch_id, await_list in self.awaited_steps.items():
            print(f"\n   Orchestration {orch_id[:16]}...")
            for await_info in await_list:
                if await_info['awaited']:
                    for awaited_id in await_info['awaited']:
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