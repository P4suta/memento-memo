export interface Memo {
	id: number;
	session_id: number;
	content: string;
	content_html: string;
	tags?: Tag[];
	pinned: boolean;
	created_at: string;
	updated_at: string;
	deleted_at?: string;
}

export interface Tag {
	id: number;
	name: string;
	memo_count?: number;
}

export interface Session {
	id: number;
	started_at: string;
	ended_at: string;
	date_label: string;
	memo_count: number;
	total_chars: number;
}

export interface HeatmapEntry {
	date_label: string;
	memo_count: number;
	total_chars: number;
}

export interface DailyReport {
	id: number;
	report_date: string;
	session_count: number;
	total_memos: number;
	total_chars: number;
	chars_per_minute: number | null;
	active_minutes: number;
	hourly_distribution: Record<string, number>;
}

export interface Stats {
	total_memos: number;
	total_sessions: number;
	active_days: number;
	total_chars: number;
}

export interface PaginatedResponse<T> {
	items: T[];
	next_cursor: string;
	has_more: boolean;
}

export interface WSMessage {
	type: string;
	payload: unknown;
	timestamp: string;
}
