import { api } from '$lib/api/client';
import type { Memo } from '$lib/api/types';

class SearchStore {
	query = $state('');
	results = $state<Memo[]>([]);
	cursor = $state<string | null>(null);
	loading = $state(false);
	hasMore = $state(false);
	private debounceTimer: ReturnType<typeof setTimeout> | null = null;

	setQuery(q: string) {
		this.query = q;
		if (this.debounceTimer) clearTimeout(this.debounceTimer);
		if (!q.trim()) {
			this.results = [];
			this.hasMore = false;
			return;
		}
		this.debounceTimer = setTimeout(() => this.search(), 300);
	}

	async search() {
		if (!this.query.trim()) return;
		this.loading = true;
		this.cursor = null;
		try {
			const res = await api.search(this.query, { limit: 20 });
			this.results = res.items;
			this.cursor = res.next_cursor || null;
			this.hasMore = res.has_more;
		} finally {
			this.loading = false;
		}
	}

	async loadMore() {
		if (this.loading || !this.hasMore || !this.cursor) return;
		this.loading = true;
		try {
			const res = await api.search(this.query, { cursor: this.cursor, limit: 20 });
			this.results = [...this.results, ...res.items];
			this.cursor = res.next_cursor || null;
			this.hasMore = res.has_more;
		} finally {
			this.loading = false;
		}
	}
}

export const searchStore = new SearchStore();
