import { api } from '$lib/api/client';
import { memoStore } from './memos.svelte';
import type { WSMessage } from '$lib/api/types';

class ReconnectingWebSocket {
	private backoff = $state(0);
	private maxBackoff = 30_000;
	private ws: WebSocket | null = null;
	private lastEventTimestamp: string | null = null;
	connected = $state(false);

	connect() {
		if (typeof window === 'undefined') return;

		const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
		this.ws = new WebSocket(`${proto}//${location.host}/api/v1/ws`);

		this.ws.onopen = () => {
			this.backoff = 0;
			this.connected = true;
			// Fetch missed messages if we have a last timestamp
			if (this.lastEventTimestamp) {
				this.fetchMissed();
			}
		};

		this.ws.onclose = () => {
			this.connected = false;
			this.scheduleReconnect();
		};

		this.ws.onerror = () => {
			this.ws?.close();
		};

		this.ws.onmessage = (e) => {
			const msg: WSMessage = JSON.parse(e.data);
			this.lastEventTimestamp = msg.timestamp;
			this.handleMessage(msg);
		};

		// Send ping every 30s
		this.startPing();
	}

	private handleMessage(msg: WSMessage) {
		switch (msg.type) {
			case 'memo.created': {
				const payload = msg.payload as { memo_id: number };
				api.getMemo(payload.memo_id).then((memo) => {
					memoStore.prependMemo(memo);
				});
				break;
			}
			case 'memo.updated': {
				const payload = msg.payload as { memo_id: number };
				api.getMemo(payload.memo_id).then((memo) => {
					memoStore.updateMemo(memo);
				});
				break;
			}
			case 'memo.deleted': {
				const payload = msg.payload as { memo_id: number };
				memoStore.removeMemo(payload.memo_id);
				break;
			}
		}
	}

	private async fetchMissed() {
		if (!this.lastEventTimestamp) return;
		try {
			const res = await api.getMemos({ since: this.lastEventTimestamp, limit: 100 });
			for (const memo of res.items) {
				memoStore.prependMemo(memo);
			}
		} catch {
			// ignore
		}
	}

	private scheduleReconnect() {
		if (document.hidden) {
			document.addEventListener('visibilitychange', () => this.connect(), { once: true });
			return;
		}
		const delay =
			this.backoff === 0 ? 0 : Math.min(1000 * 2 ** (this.backoff - 1), this.maxBackoff);
		this.backoff++;
		setTimeout(() => this.connect(), delay);
	}

	private pingInterval: ReturnType<typeof setInterval> | null = null;

	private startPing() {
		if (this.pingInterval) clearInterval(this.pingInterval);
		this.pingInterval = setInterval(() => {
			if (this.ws?.readyState === WebSocket.OPEN) {
				this.ws.send(JSON.stringify({ type: 'ping' }));
			}
		}, 30_000);
	}

	disconnect() {
		if (this.pingInterval) clearInterval(this.pingInterval);
		this.ws?.close();
	}
}

export const wsStore = new ReconnectingWebSocket();
