import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api } from './client';

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

function jsonResponse(data: unknown, status = 200) {
	return new Response(JSON.stringify(data), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
}

function errorResponse(message: string, status: number) {
	return new Response(JSON.stringify({ error: { message } }), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
}

beforeEach(() => {
	mockFetch.mockReset();
});

describe('api.createMemo', () => {
	it('sends POST with content', async () => {
		const memo = { id: 1, content: 'hello', content_html: '<p>hello</p>' };
		mockFetch.mockResolvedValueOnce(jsonResponse(memo));

		const result = await api.createMemo('hello');

		expect(mockFetch).toHaveBeenCalledOnce();
		const [url, init] = mockFetch.mock.calls[0];
		expect(url).toBe('/api/v1/memos');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ content: 'hello' });
		expect(result).toEqual(memo);
	});

	it('throws on error response', async () => {
		mockFetch.mockResolvedValueOnce(errorResponse('メモ本文を入力してください', 400));

		await expect(api.createMemo('')).rejects.toThrow('メモ本文を入力してください');
	});
});

describe('api.getMemos', () => {
	it('sends GET with query params', async () => {
		const response = { items: [], next_cursor: '', has_more: false };
		mockFetch.mockResolvedValueOnce(jsonResponse(response));

		await api.getMemos({ cursor: 'abc', limit: 20 });

		const [url] = mockFetch.mock.calls[0];
		expect(url).toContain('cursor=abc');
		expect(url).toContain('limit=20');
	});

	it('works without params', async () => {
		const response = { items: [], next_cursor: '', has_more: false };
		mockFetch.mockResolvedValueOnce(jsonResponse(response));

		const result = await api.getMemos();
		expect(result.items).toEqual([]);
	});
});

describe('api.getMemo', () => {
	it('fetches single memo by ID', async () => {
		const memo = { id: 42, content: 'test' };
		mockFetch.mockResolvedValueOnce(jsonResponse(memo));

		const result = await api.getMemo(42);

		expect(mockFetch.mock.calls[0][0]).toBe('/api/v1/memos/42');
		expect(result).toEqual(memo);
	});
});

describe('api.updateMemo', () => {
	it('sends PATCH with content', async () => {
		const memo = { id: 1, content: 'updated' };
		mockFetch.mockResolvedValueOnce(jsonResponse(memo));

		await api.updateMemo(1, 'updated');

		const [url, init] = mockFetch.mock.calls[0];
		expect(url).toBe('/api/v1/memos/1');
		expect(init.method).toBe('PATCH');
	});
});

describe('api.deleteMemo', () => {
	it('sends DELETE', async () => {
		mockFetch.mockResolvedValueOnce(new Response(null, { status: 204 }));

		await api.deleteMemo(1);

		const [url, init] = mockFetch.mock.calls[0];
		expect(url).toBe('/api/v1/memos/1');
		expect(init.method).toBe('DELETE');
	});
});

describe('api.togglePin', () => {
	it('sends POST to pin endpoint', async () => {
		const memo = { id: 1, pinned: true };
		mockFetch.mockResolvedValueOnce(jsonResponse(memo));

		const result = await api.togglePin(1);

		expect(mockFetch.mock.calls[0][0]).toBe('/api/v1/memos/1/pin');
		expect(result.pinned).toBe(true);
	});
});

describe('api.search', () => {
	it('sends query parameter', async () => {
		const response = { items: [], next_cursor: '', has_more: false };
		mockFetch.mockResolvedValueOnce(jsonResponse(response));

		await api.search('golang');

		const [url] = mockFetch.mock.calls[0];
		expect(url).toContain('q=golang');
	});
});

describe('api.getTags', () => {
	it('fetches tags list', async () => {
		const response = { items: [{ id: 1, name: 'go', memo_count: 5 }] };
		mockFetch.mockResolvedValueOnce(jsonResponse(response));

		const result = await api.getTags();
		expect(result.items).toHaveLength(1);
		expect(result.items[0].name).toBe('go');
	});
});

describe('api.getHealth', () => {
	it('fetches health status', async () => {
		const health = { status: 'ok', version: 'dev', uptime: '1h', database: 'ok' };
		mockFetch.mockResolvedValueOnce(jsonResponse(health));

		const result = await api.getHealth();
		expect(result.status).toBe('ok');
	});
});

describe('error handling', () => {
	it('falls back to HTTP status on non-JSON error', async () => {
		mockFetch.mockResolvedValueOnce(new Response('Internal Server Error', { status: 500 }));

		await expect(api.getHealth()).rejects.toThrow('HTTP 500');
	});
});
