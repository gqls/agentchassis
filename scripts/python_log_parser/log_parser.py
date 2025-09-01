#!/usr/bin/env python3
"""
Agent Communication Flow Visualizer - Enhanced with Container Tracking
"""

import json
import re
import sys
from datetime import datetime
from collections import defaultdict, OrderedDict
import subprocess

class AgentMessage:
    def __init__(self, timestamp, log_entry, source_func=None, container=None):
        self.timestamp = timestamp
        self.raw = log_entry
        self.source_func = source_func
        self.container = container  # Add container tracking

        # Extract all headers/fields
        self.headers = {}
        for key, value in log_entry.items():
            if key not in ['msg', 'ts', 'caller', 'level']:
                self.headers[key] = value

        # Core IDs
        self.correlation_id = log_entry.get('correlation_id', '')
        self.orchestration_id = log_entry.get('orchestration_id', '')
        self.request_id = log_entry.get('request_id', '')
        self.in_response_to = log_entry.get('in_response_to', '')
        self.direction = log_entry.get('direction', '')
        self.topic = log_entry.get('topic', '')
        self.message_type = log_entry.get('message_type', '')
        self.action = log_entry.get('action', '')
        self.payload_preview = log_entry.get('payload_preview', '')

        # Container/pod info
        self.container = log_entry.get('container', container)

        # Agent identification - now with explicit fields
        self.from_agent_id = log_entry.get('from_agent_id', '')
        self.from_agent_type = log_entry.get('from_agent_type', '')
        self.to_agent_id = log_entry.get('to_agent_id', '')
        self.to_agent_type = log_entry.get('to_agent_type', '')
        self.owner_agent_id = log_entry.get('owner_agent_id', '')
        self.owner_agent_type = log_entry.get('owner_agent_type', '')

        # Fallback to from/to strings if explicit fields not found
        if not self.from_agent_id or not self.from_agent_type:
            from_str = log_entry.get('from', '/')
            if '/' in from_str:
                parts = from_str.split('/')
                if not self.from_agent_id and len(parts) > 0 and parts[0]:
                    self.from_agent_id = parts[0]
                if not self.from_agent_type and len(parts) > 1 and parts[1]:
                    self.from_agent_type = parts[1]

        if not self.to_agent_id or not self.to_agent_type:
            to_str = log_entry.get('to', '/')
            if '/' in to_str:
                parts = to_str.split('/')
                if not self.to_agent_id and len(parts) > 0 and parts[0]:
                    self.to_agent_id = parts[0]
                if not self.to_agent_type and len(parts) > 1 and parts[1]:
                    self.to_agent_type = parts[1]

class FlowStep:
    """Represents a step in the execution flow"""
    def __init__(self, timestamp, step_type, details, source_func, container=None):
        self.timestamp = timestamp
        self.step_type = step_type
        self.details = details
        self.source_func = source_func
        self.container = container

class LogParser:
    def __init__(self):
        self.messages = []
        self.agents = {}  # agent_id -> agent_info
        self.containers = {}  # container -> agent mapping
        self.orchestrations = defaultdict(list)
        self.awaited_steps = defaultdict(list)
        self.spawned_agents = {}
        self.flow_steps = []
        self.request_mappings = {}
        self.orchestrator_agent = None

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
            # Extract pod/container name from prefix
            container = None
            if line.startswith('[pod/'):
                match = re.match(r'\[pod/([^/\]]+)(?:/([^\]]+))?\]', line)
                if match:
                    pod_name = match.group(1)
                    container_name = match.group(2) if match.group(2) else 'agent-chassis'
                    # Extract short container ID (last part after last dash)
                    container = pod_name.split('-')[-1] if pod_name else None

            json_match = re.search(r'\{.*\}$', line)
            if not json_match:
                return

            log_entry = json.loads(json_match.group())

            timestamp = log_entry.get('ts', '')
            caller = log_entry.get('caller', '')
            source_func = caller.split('/')[-1] if caller else 'unknown'
            msg = log_entry.get('msg', '')

            # Track agent spawning with container info
            if 'Successfully spawned agent job' in msg:
                agent_id = log_entry.get('DEBUG_SPAWN_31: agent_id', '')
                agent_type = log_entry.get('DEBUG_SPAWN_31: agent_type', '')
                job_name = log_entry.get('DEBUG_SPAWN_31: job_name', '')

                if agent_id and agent_type:
                    # Extract container ID from job name (last 8 chars of agent ID)
                    container_id = agent_id[:8] if agent_id else ''

                    self.agents[agent_id] = {
                        'type': agent_type,
                        'id': agent_id[:8],
                        'full_id': agent_id,
                        'job_name': job_name,
                        'container': container_id,
                        'spawned_from': container
                    }
                    print(f"   🤖 Spawned: {agent_type} [{agent_id[:8]}] in job {job_name} from container [{container}]")

            # Track agent roles
            elif 'Successfully spawned agent' in msg:
                agent_id = log_entry.get('DEBUG_SPAWN_12: agent_id', '')
                agent_type = log_entry.get('DEBUG_SPAWN_12: agent_type', '')
                role = log_entry.get('DEBUG_SPAWN_12: role', '')

                if agent_id and agent_id in self.agents:
                    self.agents[agent_id]['role'] = role
                    if role == 'orchestrator':
                        self.orchestrator_agent = agent_id
                        print(f"   📋 Role assigned: {agent_type} ({role}) [{agent_id[:8]}]")

            # Track pre-generated request IDs
            elif 'Pre-generated request ID' in msg or 'CRITICAL_FLOW: Pre-generating request ID' in msg:
                original = log_entry.get('original_request_id', '') or log_entry.get('request_id', '')
                generated = log_entry.get('pre_generated_request_id', '') or log_entry.get('action_request_id', '')
                if original and generated and original != generated:
                    self.request_mappings[original] = generated
                    self.flow_steps.append(FlowStep(
                        timestamp,
                        "PRE_GENERATE_REQUEST",
                        {
                            'original': original[:16],
                            'generated': generated[:16],
                            'action': log_entry.get('action', ''),
                            'step': log_entry.get('step', ''),
                            'container': container
                        },
                        source_func,
                        container
                    ))

            # Track awaited steps
            elif msg == 'AWAITED_STEPS_CHANGED' or 'Successfully added request to awaited steps' in msg:
                orch_id = log_entry.get('orchestration_id', '')
                awaited = log_entry.get('awaited_steps', []) or log_entry.get('all_awaited_steps', [])
                action = log_entry.get('for_action', '') or log_entry.get('action', '')

                if orch_id and awaited:
                    self.awaited_steps[orch_id].append({
                        'timestamp': timestamp,
                        'awaited': awaited,
                        'action': action,
                        'request_id': log_entry.get('request_id', ''),
                        'container': container
                    })
                    self.flow_steps.append(FlowStep(
                        timestamp,
                        "AWAIT_SETUP",
                        {
                            'orchestration_id': orch_id[:16],
                            'awaited_requests': [r[:16] for r in awaited],
                            'action': action,
                            'container': container
                        },
                        source_func,
                        container
                    ))

            # Track MESSAGE_TRACE entries
            elif msg == 'MESSAGE_TRACE':
                msg_obj = AgentMessage(timestamp, log_entry, source_func, container)

                # Map container to agent if possible
                if msg_obj.owner_agent_id and container:
                    self.containers[container] = msg_obj.owner_agent_id

                self.messages.append(msg_obj)

                if msg_obj.orchestration_id:
                    self.orchestrations[msg_obj.orchestration_id].append(msg_obj)

            # Track response sending
            elif 'TRACE: Sending workflow response' in msg or 'CRITICAL_FLOW: sendWorkflowResponse called' in msg:
                self.flow_steps.append(FlowStep(
                    timestamp,
                    "SEND_RESPONSE",
                    {
                        'orchestration_id': log_entry.get('orchestration_id', '')[:16] or log_entry.get('response_orch_id', '')[:16],
                        'in_response_to': (log_entry.get('in_response_to', '') or 'none')[:16],
                        'original_request_id': log_entry.get('original_request_id', '')[:16] if log_entry.get('original_request_id') else '',
                        'container': container
                    },
                    source_func,
                    container
                ))

            # Track waiting states
            elif 'Execution paused - waiting for responses' in msg or 'Step resulted in waiting state' in msg:
                self.flow_steps.append(FlowStep(
                    timestamp,
                    "EXECUTION_PAUSED",
                    {
                        'orchestration_id': log_entry.get('orchestration_id', '')[:16],
                        'status': log_entry.get('status', ''),
                        'step': log_entry.get('step', ''),
                        'container': container
                    },
                    source_func,
                    container
                ))

        except Exception as e:
            pass

    def format_agent_display(self, agent_id, agent_type, container=None):
        """Format agent for display with type, container, and ID"""
        if agent_id and agent_id in self.agents:
            info = self.agents[agent_id]
            role_str = f"({info.get('role')})" if info.get('role') else ""
            container_str = f"@{info.get('container', '?')[:5]}" if info.get('container') else ""
            return f"{info['type']}{role_str}[{info['id']}]{container_str}"
        elif agent_type:
            short_id = agent_id[:8] if agent_id and len(agent_id) > 8 else agent_id or "?"
            container_str = f"@{container[:5]}" if container else ""
            return f"{agent_type}[{short_id}]{container_str}"
        elif agent_id:
            short_id = agent_id[:8] if len(agent_id) > 8 else agent_id
            container_str = f"@{container[:5]}" if container else ""
            return f"agent[{short_id}]{container_str}"
        else:
            return "SYSTEM"

    def print_message(self, msg):
        """Print a single message with container info"""
        # Determine message type indicator
        indicators = {
            'sending_response': '→ RESPONSE',
            'processing_response': '← PROCESSING',
            'received': '↓ RECEIVED',
            'sending': '→ SENDING',
            'awaited_update': '⏳ AWAITED',
        }
        indicator = indicators.get(msg.direction, '↔ MESSAGE')

        # Color codes
        colors = {
            'sending_response': '\033[92m',  # Green
            'processing_response': '\033[94m',  # Blue
            'received': '\033[93m',  # Yellow
            'sending': '\033[96m',  # Cyan
            'awaited_update': '\033[95m',  # Purple
        }
        color = colors.get(msg.direction, '\033[90m')

        from_str = self.format_agent_display(msg.from_agent_id, msg.from_agent_type, msg.container)
        to_str = self.format_agent_display(msg.to_agent_id, msg.to_agent_type, msg.container)

        print(f"\n{color}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        print(f"[{msg.source_func}] {msg.timestamp}")
        if msg.container:
            print(f"📦 Container: {msg.container}")
        print(f"{indicator}: {from_str} → {to_str}")

        # Print key headers
        print(f"\n📋 Key Headers:")
        important_headers = ['correlation_id', 'orchestration_id', 'request_id', 'in_response_to',
                             'owner_agent_id', 'owner_agent_type', 'topic']
        for key in important_headers:
            if key in msg.headers and msg.headers[key]:
                value = msg.headers[key]
                if len(str(value)) > 50:
                    value = f"{str(value)[:50]}..."
                print(f"   {key}: {value}")

        # Parse and show payload if available
        if msg.payload_preview:
            print(f"\n📦 Payload:")
            try:
                if isinstance(msg.payload_preview, str):
                    if msg.payload_preview.isdigit():
                        print(f"   Size: {msg.payload_preview} bytes")
                    else:
                        payload = json.loads(msg.payload_preview)
                        print(f"   {json.dumps(payload, indent=3)[:500]}")
                else:
                    print(f"   {str(msg.payload_preview)[:500]}")
            except:
                print(f"   {msg.payload_preview[:200]}...")

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
                'EXECUTION_PAUSED': '\033[91m',  # Red
            }.get(step.step_type, '\033[90m')

            print(f"{color}[{step.source_func}] {step.timestamp}")
            if step.container:
                print(f"   📦 Container: {step.container}")
            print(f"   {step.step_type}")

            for key, value in step.details.items():
                if value and key != 'container':  # Don't repeat container
                    print(f"      {key}: {value}")
            print('\033[0m')

    def print_flow(self, correlation_id=None):
        """Print message flow with enhanced details"""
        self.print_flow_trace()

        messages = self.messages
        if correlation_id:
            messages = [m for m in messages if m.correlation_id == correlation_id]

        print(f"\n{'='*100}")
        print(f"Agent Communication Flow")
        print(f"{'='*100}\n")

        if self.request_mappings:
            print("🔗 Request ID Mappings:")
            for orig, gen in self.request_mappings.items():
                print(f"   {orig[:16]}... → {gen[:16]}... (pre-generated)")
            print()

        for orch_id, orch_messages in self.orchestrations.items():
            if correlation_id and not any(m.correlation_id == correlation_id for m in orch_messages):
                continue

            print(f"📋 Orchestration: {orch_id[:16]}...")

            if orch_id in self.awaited_steps:
                print(f"   ⏳ Awaited Steps:")
                for await_info in self.awaited_steps[orch_id]:
                    container_str = f" (from container {await_info.get('container')})" if await_info.get('container') else ""
                    print(f"      - Action: {await_info['action']}{container_str}")
                    for req_id in await_info['awaited']:
                        print(f"        Request ID: {req_id[:16]}...")

            print(f"\n   Messages:")
            for msg in sorted(orch_messages, key=lambda x: x.timestamp):
                self.print_message(msg)

    def print_summary(self):
        """Print enhanced summary with container info"""
        print(f"\n{'='*100}")
        print("Summary Analysis")
        print(f"{'='*100}\n")

        print(f"Total Messages: {len(self.messages)}")
        print(f"Total Agents Tracked: {len(self.agents)}")
        print(f"Total Containers: {len(self.containers)}")
        print(f"Total Orchestrations: {len(self.orchestrations)}")

        if self.agents:
            print("\n🤖 Agent Deployment Map:")
            for agent_id, info in self.agents.items():
                role = f"({info.get('role')})" if info.get('role') else ""
                marker = " ⭐" if agent_id == self.orchestrator_agent else ""
                print(f"   {info['type']} {role}{marker}")
                print(f"      ID: {agent_id[:16]}...")
                print(f"      Job: {info.get('job_name', 'unknown')}")
                print(f"      Container: {info.get('container', 'unknown')}")
                if info.get('spawned_from'):
                    print(f"      Spawned from: {info.get('spawned_from')}")

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