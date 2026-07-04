#!/usr/bin/env python3
"""netmetrics - Scapy tabanli paket yakalama ajani."""

from __future__ import annotations

import argparse
import json
import signal
import socket
import sys
import threading
import time
from collections import Counter

from scapy.all import ARP, ICMP, IP, IPv6, TCP, UDP, AsyncSniffer


class Metrics:
    """Paket metriklerini thread-guvenli sekilde toplar."""

    def __init__(self, top_n: int = 5) -> None:
        self.top_n = top_n
        self._lock = threading.Lock()
        self._total_packets = 0
        self._total_bytes = 0
        self._reset_window()

    def _reset_window(self) -> None:
        self._win_packets = 0
        self._win_bytes = 0
        self._proto = Counter()
        self._src = Counter()
        self._dst = Counter()

    def process(self, pkt) -> None:
        size = len(pkt)
        with self._lock:
            self._win_packets += 1
            self._win_bytes += size
            self._total_packets += 1
            self._total_bytes += size

            if pkt.haslayer(TCP):
                self._proto["tcp"] += 1
            elif pkt.haslayer(UDP):
                self._proto["udp"] += 1
            elif pkt.haslayer(ICMP):
                self._proto["icmp"] += 1
            elif pkt.haslayer(ARP):
                self._proto["arp"] += 1
            else:
                self._proto["other"] += 1

            if pkt.haslayer(IP):
                self._src[pkt[IP].src] += 1
                self._dst[pkt[IP].dst] += 1
            elif pkt.haslayer(IPv6):
                self._src[pkt[IPv6].src] += 1
                self._dst[pkt[IPv6].dst] += 1

    def snapshot(self) -> dict:
        with self._lock:
            snap = {
                "window": {
                    "packets": self._win_packets,
                    "bytes": self._win_bytes,
                    "protocols": dict(self._proto),
                    "top_talkers": dict(self._src.most_common(self.top_n)),
                    "top_destinations": dict(self._dst.most_common(self.top_n)),
                },
                "lifetime": {
                    "packets": self._total_packets,
                    "bytes": self._total_bytes,
                },
            }
            self._reset_window()
        return snap


class UDPPusher:
    def __init__(self, host: str, port: int) -> None:
        self._addr = (host, port)
        self._sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

    def push(self, payload: dict) -> int:
        data = json.dumps(payload, separators=(",", ":")).encode("utf-8")
        if len(data) > 65_507:
            raise ValueError(f"payload too large: {len(data)} bytes")
        return self._sock.sendto(data, self._addr)

    def close(self) -> None:
        self._sock.close()


def parse_args(argv=None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Scapy tabanli ag metrik ajani")
    p.add_argument("--iface", default=None, help="Dinlenecek arayuz (bos: otomatik)")
    p.add_argument("--filter", dest="bpf", default=None, help="BPF filtresi")
    p.add_argument("--remote-host", default="127.0.0.1", help="Uzak sunucu adresi")
    p.add_argument("--remote-port", type=int, default=9999, help="Uzak UDP portu")
    p.add_argument("--interval", type=float, default=5.0, help="Push araligi (sn)")
    p.add_argument("--agent-id", default=socket.gethostname(), help="Ajan kimligi")
    p.add_argument("--top-n", type=int, default=5, help="Top-talker liste boyutu")
    return p.parse_args(argv)


def main(argv=None) -> int:
    args = parse_args(argv)
    metrics = Metrics(top_n=args.top_n)
    pusher = UDPPusher(args.remote_host, args.remote_port)

    sniffer = AsyncSniffer(
        iface=args.iface,
        filter=args.bpf,
        prn=metrics.process,
        store=False,
    )
    sniffer.start()
    print(
        f"[netmetrics] capture basladi -> {args.remote_host}:{args.remote_port} "
        f"(iface={args.iface or 'auto'}, interval={args.interval}s)",
        flush=True,
    )

    running = {"on": True}

    def _stop(signum, frame):
        running["on"] = False

    signal.signal(signal.SIGINT, _stop)
    signal.signal(signal.SIGTERM, _stop)

    try:
        while running["on"]:
            time.sleep(args.interval)
            payload = {
                "agent_id": args.agent_id,
                "timestamp": time.time(),
                "interval": args.interval,
                "metrics": {"packet_capture": metrics.snapshot()},
            }
            try:
                sent = pusher.push(payload)
                pkts = payload["metrics"]["packet_capture"]["window"]["packets"]
                print(f"[netmetrics] push: {pkts} paket, {sent} bayt", flush=True)
            except Exception as exc:
                print(f"[netmetrics] push hatasi: {exc}", flush=True)
    finally:
        sniffer.stop()
        pusher.close()
        print("[netmetrics] durdu.", flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())