import type { Memo, Tag, Session, HeatmapEntry, DailyReport, Stats, PaginatedResponse } from './types';

const BASE = '/api/v1';

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
	const res = await fetch(url, {
		headers: { 'Content-Type': 'application/json', ...init?.headers },
		...init
	});
	if (!res.ok) {
		const body = await res.json().catch(() => ({}));
		throw new Error(body?.error?.message || `HTTP ${res.status}`);
	}
	if (res.status === 204) return undefined as T;
	return res.json();
}

export const api = {
	// Memos
	createMemo(content: string) {
		return fetchJSON<Memo>(`${BASE}/memos`, {
			method: 'POST',
			body: JSON.stringify({ content })
		});
	},

	getMemos(params?: { cursor?: string; limit?: number; since?: string }) {
		const q = new URLSearchParams();
		if (params?.cursor) q.set('cursor', params.cursor);
		if (params?.limit) q.set('limit', String(params.limit));
		if (params?.since) q.set('since', params.since);
		return fetchJSON<PaginatedResponse<Memo>>(`${BASE}/memos?${q}`);
	},

	getMemo(id: number) {
		return fetchJSON<Memo>(`${BASE}/memos/${id}`);
	},

	updateMemo(id: number, content: string) {
		return fetchJSON<Memo>(`${BASE}/memos/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ content })
		});
	},

	deleteMemo(id: number) {
		return fetchJSON<void>(`${BASE}/memos/${id}`, { method: 'DELETE' });
	},

	restoreMemo(id: number) {
		return fetchJSON<void>(`${BASE}/memos/${id}/restore`, { method: 'POST' });
	},

	permanentDeleteMemo(id: number) {
		return fetchJSON<void>(`${BASE}/memos/${id}/permanent`, { method: 'DELETE' });
	},

	togglePin(id: number) {
		return fetchJSON<Memo>(`${BASE}/memos/${id}/pin`, { method: 'POST' });
	},

	// Search
	search(q: string, params?: { cursor?: string; limit?: number }) {
		const query = new URLSearchParams({ q });
		if (params?.cursor) query.set('cursor', params.cursor);
		if (params?.limit) query.set('limit', String(params.limit));
		return fetchJSON<PaginatedResponse<Memo>>(`${BASE}/search?${query}`);
	},

	// Tags
	getTags() {
		return fetchJSON<{ items: Tag[] }>(`${BASE}/tags`);
	},

	getTagMemos(name: string, params?: { cursor?: string; limit?: number }) {
		const q = new URLSearchParams();
		if (params?.cursor) q.set('cursor', params.cursor);
		if (params?.limit) q.set('limit', String(params.limit));
		return fetchJSON<PaginatedResponse<Memo>>(`${BASE}/tags/${encodeURIComponent(name)}/memos?${q}`);
	},

	// Sessions
	getSessions(from: string, to: string) {
		return fetchJSON<{ items: Session[] }>(`${BASE}/sessions?from=${from}&to=${to}`);
	},

	getHeatmap(year?: number) {
		const q = year ? `?year=${year}` : '';
		return fetchJSON<{ items: HeatmapEntry[] }>(`${BASE}/calendar/heatmap${q}`);
	},

	getDaily(date: string) {
		return fetchJSON<{ date: string; sessions: Session[]; memos: Memo[] }>(
			`${BASE}/calendar/daily/${date}`
		);
	},

	// Reports
	getDailyReport(date: string) {
		return fetchJSON<DailyReport>(`${BASE}/reports/daily/${date}`);
	},

	getStats() {
		return fetchJSON<Stats>(`${BASE}/reports/stats`);
	},

	// Health
	getHealth() {
		return fetchJSON<{ status: string; version: string; uptime: string; database: string }>(
			`${BASE}/health`
		);
	}
};
