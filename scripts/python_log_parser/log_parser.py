#!/usr/bin/env python3
"""
Agent Communication Flow Visualizer
Parses Kubernetes logs and creates a visual diagram of agent communications
"""

import json
import re
import sys
from datetime import datetime
from collections import defaultdict
import subprocess

# Try to import optional dependencies
try:
    import matplotlib.pyplot as plt
    import matplotlib.patches as mpatches
    from matplotlib.patches import FancyBboxPatch, FancyArrowPatch
    import networkx as nx
    HAVE_MATPLOTLIB = True
except ImportError:
    HAVE_MATPLOTLIB = False
    print("Warning: matplotlib not found. Install with: pip install matplotlib networkx")

class AgentMessage:
    def __init__(self, timestamp, action, headers, payload=None):
        self.timestamp = timestamp
        self.action = action
        self.headers = headers
        self.payload = payload
        self.correlation_id = headers.get('correlation_id', '')
        self.orchestration_id = headers.get('orchestration_id', '')
        self.from_agent = headers.get('from_agent_id', '')
        self.from_type = headers.get('from_agent_type', '')
        self.to_agent = headers.get('to_agent_id', '')
        self.to_type = headers.get('to_agent_type', '')
        self.message_type = headers.get('message_type', '')
        self.request_id = headers.get('request_id', '')
        self.in_response_to = headers.get('in_response_to', '')

class LogParser:
    def __init__(self):
        self.messages = []
        self.agents = set()
        self.orchestrations = defaultdict(list)

    def parse_k8s_logs(self, namespace="ai-persona-system", label="app=agent-chassis"):
        """Fetch and parse Kubernetes logs"""
        cmd = f"kubectl -n {namespace} logs -l {label} --prefix=true --tail=1000"

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
        # Look for MESSAGE_TRACE entries
        if 'MESSAGE_TRACE' not in line:
            return

        try:
            # Extract JSON part of the log
            json_match = re.search(r'\{.*\}$', line)
            if not json_match:
                return

            log_entry = json.loads(json_match.group())

            # Extract headers from the log
            headers = {}
            if 'all_headers' in log_entry:
                headers = log_entry['all_headers']
            else:
                # Build headers from individual fields
                for key in ['correlation_id', 'orchestration_id', 'message_id',
                           'request_id', 'message_type', 'from_agent_id',
                           'from_agent_type', 'to_agent_id', 'to_agent_type',
                           'in_response_to']:
                    if key in log_entry:
                        headers[key] = log_entry[key]

            # Parse payload if present
            payload = None
            if 'payload' in log_entry:
                try:
                    payload = json.loads(log_entry['payload'])
                except:
                    payload = log_entry['payload']

            # Create message object
            msg = AgentMessage(
                timestamp=log_entry.get('ts', ''),
                action=log_entry.get('action', ''),
                headers=headers,
                payload=payload
            )

            self.messages.append(msg)

            # Track agents
            if msg.from_agent:
                self.agents.add((msg.from_agent, msg.from_type))
            if msg.to_agent:
                self.agents.add((msg.to_agent, msg.to_type))

            # Group by orchestration
            if msg.orchestration_id:
                self.orchestrations[msg.orchestration_id].append(msg)

        except Exception as e:
            # Skip malformed lines
            pass

    def print_flow(self, correlation_id=None):
        """Print message flow in text format"""
        messages = self.messages
        if correlation_id:
            messages = [m for m in messages if m.correlation_id == correlation_id]

        print(f"\n{'='*80}")
        print(f"Agent Communication Flow")
        if correlation_id:
            print(f"Correlation ID: {correlation_id}")
        print(f"{'='*80}\n")

        for msg in sorted(messages, key=lambda x: x.timestamp):
            # Format the message
            if msg.message_type == 'request':
                arrow = "→"
                color = '\033[92m'  # Green
            elif msg.message_type == 'response':
                arrow = "←"
                color = '\033[94m'  # Blue
            else:
                arrow = "↔"
                color = '\033[93m'  # Yellow

            from_str = f"{msg.from_type}[{msg.from_agent[:8]}]" if msg.from_agent else "UNKNOWN"
            to_str = f"{msg.to_type}[{msg.to_agent[:8]}]" if msg.to_agent else "UNKNOWN"

            print(f"{color}{msg.timestamp}")
            print(f"  {from_str} {arrow} {to_str}")
            print(f"  Action: {msg.action}")
            print(f"  Type: {msg.message_type}")

            if msg.request_id:
                print(f"  Request ID: {msg.request_id}")
            if msg.in_response_to:
                print(f"  In Response To: {msg.in_response_to}")

            if msg.payload:
                print(f"  Payload: {json.dumps(msg.payload, indent=4)[:200]}...")

            print('\033[0m')  # Reset color
            print()

    def visualize_flow(self, correlation_id=None):
        """Create a visual diagram of message flow"""
        if not HAVE_MATPLOTLIB:
            print("Matplotlib not available. Install with: pip install matplotlib networkx")
            return

        messages = self.messages
        if correlation_id:
            messages = [m for m in messages if m.correlation_id == correlation_id]

        if not messages:
            print("No messages to visualize")
            return

        # Create directed graph
        G = nx.DiGraph()

        # Add nodes for agents
        agent_positions = {}
        agent_types = {}
        for i, (agent_id, agent_type) in enumerate(self.agents):
            short_id = agent_id[:8] if agent_id else 'unknown'
            node_label = f"{agent_type}\n{short_id}"
            G.add_node(node_label)
            agent_positions[agent_id] = node_label
            agent_types[agent_id] = agent_type

        # Add edges for messages
        edge_labels = {}
        for msg in messages:
            if msg.from_agent and msg.to_agent:
                from_node = agent_positions.get(msg.from_agent, 'unknown')
                to_node = agent_positions.get(msg.to_agent, 'unknown')

                # Add edge with message details
                G.add_edge(from_node, to_node)

                # Create edge label
                label = f"{msg.action}\n{msg.message_type}"
                if msg.request_id:
                    label += f"\n{msg.request_id[:8]}"
                edge_labels[(from_node, to_node)] = label

        # Create visualization
        plt.figure(figsize=(15, 10))
        pos = nx.spring_layout(G, k=2, iterations=50)

        # Draw nodes
        nx.draw_networkx_nodes(G, pos, node_size=3000, node_color='lightblue',
                              node_shape='o', alpha=0.9)

        # Draw edges
        nx.draw_networkx_edges(G, pos, edge_color='gray',
                              connectionstyle='arc3,rad=0.1',
                              arrowsize=20, alpha=0.7)

        # Draw labels
        nx.draw_networkx_labels(G, pos, font_size=8, font_weight='bold')
        nx.draw_networkx_edge_labels(G, pos, edge_labels, font_size=6)

        plt.title(f"Agent Communication Flow\n{correlation_id if correlation_id else 'All Messages'}")
        plt.axis('off')
        plt.tight_layout()
        plt.show()

    def print_summary(self):
        """Print summary statistics"""
        print(f"\n{'='*80}")
        print("Summary Statistics")
        print(f"{'='*80}\n")

        print(f"Total Messages: {len(self.messages)}")
        print(f"Total Agents: {len(self.agents)}")
        print(f"Total Orchestrations: {len(self.orchestrations)}")

        # Count by message type
        type_counts = defaultdict(int)
        for msg in self.messages:
            type_counts[msg.message_type] += 1

        print("\nMessage Types:")
        for msg_type, count in type_counts.items():
            print(f"  {msg_type}: {count}")

        # Count by action
        action_counts = defaultdict(int)
        for msg in self.messages:
            action_counts[msg.action] += 1

        print("\nTop Actions:")
        for action, count in sorted(action_counts.items(), key=lambda x: x[1], reverse=True)[:10]:
            print(f"  {action}: {count}")

        # List orchestrations
        print("\nOrchestrations:")
        for orch_id, msgs in self.orchestrations.items():
            print(f"  {orch_id}: {len(msgs)} messages")

def main():
    parser = LogParser()

    # Parse logs from Kubernetes
    print("Fetching logs from Kubernetes...")
    parser.parse_k8s_logs()

    if not parser.messages:
        print("No MESSAGE_TRACE entries found in logs.")
        print("Make sure to add detailed logging to your coordinator.")
        return

    # Print summary
    parser.print_summary()

    # Print flow for all messages
    parser.print_flow()

    # If we have matplotlib, create visualization
    if HAVE_MATPLOTLIB:
        print("\nGenerating visualization...")
        parser.visualize_flow()

    # Offer to filter by correlation ID
    if parser.orchestrations:
        print("\nAvailable correlation IDs:")
        for i, (orch_id, msgs) in enumerate(parser.orchestrations.items()):
            if i < 10:  # Show first 10
                print(f"  {orch_id}")

        correlation_id = input("\nEnter correlation ID to filter (or press Enter for all): ").strip()
        if correlation_id:
            parser.print_flow(correlation_id)
            if HAVE_MATPLOTLIB:
                parser.visualize_flow(correlation_id)

if __name__ == "__main__":
    main()