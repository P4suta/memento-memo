import { api } from '$lib/api/client';
import type { Memo } from '$lib/api/types';

class MemoStore {
	items = $state<Memo[]>([]);
	cursor = $state<string | null>(null);
	loading = $state(false);
	hasMore = $state(true);

	async loadMore() {
		if (this.loading || !this.hasMore) return;
		this.loading = true;
		try {
			const params: { cursor?: string; limit: number } = { limit: 20 };
			if (this.cursor) params.cursor = this.cursor;
			const res = await api.getMemos(params);
			this.items = [...this.items, ...res.items];
			this.cursor = res.next_cursor || null;
			this.hasMore = res.has_more;
		} finally {
			this.loading = false;
		}
	}

	async create(content: string) {
		const memo = await api.createMemo(content);
		this.items = [memo, ...this.items];
		return memo;
	}

	async update(id: number, content: string) {
		const memo = await api.updateMemo(id, content);
		this.items = this.items.map((m) => (m.id === id ? memo : m));
		return memo;
	}

	async remove(id: number) {
		await api.deleteMemo(id);
		this.items = this.items.filter((m) => m.id !== id);
	}

	async restore(id: number) {
		await api.restoreMemo(id);
	}

	async togglePin(id: number) {
		const memo = await api.togglePin(id);
		this.items = this.items.map((m) => (m.id === id ? memo : m));
	}

	prependMemo(memo: Memo) {
		if (!this.items.find((m) => m.id === memo.id)) {
			this.items = [memo, ...this.items];
		}
	}

	updateMemo(memo: Memo) {
		this.items = this.items.map((m) => (m.id === memo.id ? memo : m));
	}

	removeMemo(id: number) {
		this.items = this.items.filter((m) => m.id !== id);
	}

	reset() {
		this.items = [];
		this.cursor = null;
		this.hasMore = true;
	}
}

export const memoStore = new MemoStore();
