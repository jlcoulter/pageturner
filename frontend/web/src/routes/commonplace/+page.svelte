<script lang="ts">
	import { onMount } from 'svelte';

	type ApiBook = {
		ID: string;
		Book: string;
		Rating: number;
		StartDate: { Time: string; Valid: boolean };
		FinishDate: { Time: string; Valid: boolean };
		Pages: { Int32: number; Valid: boolean };
		Thoughts: { String: string; Valid: boolean };
	};

	type Entry = {
		id: string;
		book: string;
		rating: number;
		startDate: string | null;
		finishDate: string | null;
		pages: number | null;
		thoughts: string | null;
	};

	function formatDate(iso: string | null): string {
		if (!iso) return '-';
		const d = new Date(iso);
		const day = String(d.getDate()).padStart(2, '0');
		const month = String(d.getMonth() + 1).padStart(2, '0');
		const year = d.getFullYear();
		return `${day}-${month}-${year}`;
	}

	let entries: Entry[] = [];
	let expanded: Set<string> = new Set();
	let loading = true;
	let error: string | null = null;

	function toggleExpand(id: string) {
		if (expanded.has(id)) {
			expanded.delete(id);
		} else {
			expanded.add(id);
		}
		expanded = expanded; // trigger reactivity
	}

	onMount(async () => {
		try {
			const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
			const res = await fetch(`${apiUrl}/api/books`);
			if (!res.ok) throw new Error(`Request failed: ${res.status}`);

			const data: ApiBook[] = await res.json();

			entries = data.map((b) => ({
				id: b.ID,
				book: b.Book,
				rating: b.Rating,
				startDate: b.StartDate?.Valid ? b.StartDate.Time : null,
				finishDate: b.FinishDate?.Valid ? b.FinishDate.Time : null,
				pages: b.Pages?.Valid ? b.Pages.Int32 : null,
				thoughts: b.Thoughts?.Valid ? b.Thoughts.String : null
			}));
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});
</script>

<div class="w-full px-6 py-6 pr-2 pl-20">
	<h1 class="mb-6 text-3xl font-bold">Commonplace Book</h1>

	<div class="grid grid-cols-3 gap-2">
		{#each entries as entry}
			<article class="flex flex-col rounded-lg border bg-base-100 p-4 shadow-sm">
				<h2 class="mb-2 text-lg font-semibold">{entry.book}</h2>

				<div
					class="prose max-w-none flex-1 {expanded.has(entry.id) ? '' : 'line-clamp-6'}"
				>
					{@html entry.thoughts}
				</div>

				{#if entry.thoughts && entry.thoughts.length > 300}
					<button
						class="btn btn-ghost btn-xs mt-1 self-start"
						on:click={() => toggleExpand(entry.id)}
					>
						{expanded.has(entry.id) ? 'Show less' : 'Read more'}
					</button>
				{/if}

				<div class="mt-3 flex flex-col gap-1 text-sm opacity-70">
					<div class="flex gap-4">
						<span>🩷 {entry.rating}/10</span>
						<span>{entry.pages} pages</span>
					</div>
					<span>
						{formatDate(entry.startDate)} → {formatDate(entry.finishDate)}
					</span>
				</div>
			</article>
		{/each}
	</div>
</div>