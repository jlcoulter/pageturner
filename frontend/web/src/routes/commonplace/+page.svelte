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

	let entries: Entry[] = [];
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		try {
			const res = await fetch('http://localhost:8080/api/books');
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

	<div class="grid grid-cols-[repeat(auto-fit,minmax(420px,1fr))] justify-items-center gap-6">
		{#each entries as entry}
			<article class="flex flex-col rounded-lg border bg-base-100 p-4 shadow-sm">
				<h2 class="mb-2 text-lg font-semibold">{entry.book}</h2>

				<div class="prose max-w-none flex-1">
					{@html entry.thoughts}
				</div>

				<div class="mt-3 flex items-center justify-between text-sm opacity-70">
					<div class="flex gap-4">
						<span>🩷 {entry.rating}/10</span>
						<span>{entry.pages} pages</span>
					</div>
					<span>
						{entry.startDate} → {entry.finishDate}
					</span>
				</div>
			</article>
		{/each}
	</div>
</div>
