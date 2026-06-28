<script lang="ts">
	import 'prosekit/basic/style.css';
	import 'prosekit/basic/typography.css';

	let selectedBook = '';
	let manualEntry = false;
	let manualTitle = '';
	let manualAuthor = '';
	let rating = 0;
	let startDate = '';
	let finishDate = '';
	let pages = '';
	$: pages = pages.replace(/[^0-9]/g, '');
	let thoughts = '';

	$: bookTitle = manualEntry
		? manualAuthor.trim()
			? `${manualTitle.trim()} by ${manualAuthor.trim()}`
			: manualTitle.trim()
		: selectedBook || searchTerm;

	async function submit() {
		const data = {
			book: bookTitle,
			rating,
			startDate,
			finishDate,
			pages,
			thoughts
		};

		try {
			const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
			const res = await fetch(`${apiUrl}/api/book`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(data)
			});

			if (!res.ok) {
				const text = await res.text();
				console.error('Failed to save book:', text);
				return;
			}

			const result = await res.json();
			console.log('Book entry saved successfully:', result);

			// Clear form
			selectedBook = '';
			searchTerm = '';
			manualTitle = '';
			manualAuthor = '';
			manualEntry = false;
			rating = 0;
			startDate = '';
			finishDate = '';
			pages = '';
			thoughts = '';
		} catch (err) {
			console.error('Error saving book entry:', err);
		}
	}

	// Search functionality
	let searchTerm = '';
	type OpenLibraryBook = { work_title: string; author_name: string; work_id: string };
	let searchResults: OpenLibraryBook[] = [];
	let searchTimeout: NodeJS.Timeout;
	let loading = false;
	let searchPerformed = false;

	async function handleSearch() {
		clearTimeout(searchTimeout);
		loading = true;
		manualEntry = false;

		searchTimeout = setTimeout(async () => {
			const term = searchTerm.trim();

			if (!term) {
				searchResults = [];
				searchPerformed = false;
				loading = false;
				return;
			}

			try {
				const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
				const res = await fetch(`${apiUrl}/api/openlibrary/search`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({ term, searchBy: 'both' })
				});

				if (!res.ok) {
					console.error('Search failed:', await res.text());
					searchResults = [];
					searchPerformed = true;
					loading = false;
					return;
				}

				const results: OpenLibraryBook[] = await res.json();

				searchResults = results.map((r: any) => ({
					work_id: r.WorkID,
					work_title: r.Title?.Valid ? r.Title.String : 'Untitled',
					author_name: r.AuthorNames?.Valid ? r.AuthorNames.String : 'Unknown'
				}));
				searchPerformed = true;
			} catch (err) {
				console.error('Error during search:', err);
				searchResults = [];
				searchPerformed = true;
			} finally {
				loading = false;
			}
		}, 300);
	}

	function selectBook(book: OpenLibraryBook) {
		selectedBook = book.work_title;
		searchTerm = book.work_title;
		searchResults = [];
		searchPerformed = false;
	}

	function enterManually() {
		manualEntry = true;
		manualTitle = searchTerm;
		searchResults = [];
		searchPerformed = false;
	}

	function cancelManual() {
		manualEntry = false;
		manualTitle = '';
		manualAuthor = '';
	}
</script>

<div class="mx-auto max-w-3xl space-y-6">
	<h1 class="text-3xl font-bold">Record a Book</h1>

	{#if manualEntry}
		<!-- Manual entry mode -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h2 class="text-lg font-semibold">Enter book manually</h2>
				<button class="btn btn-sm btn-ghost" on:click={cancelManual}>Cancel</button>
			</div>
			<div class="form-control w-full">
				<label class="label"><span class="label-text">Title</span></label>
				<input
					type="text"
					bind:value={manualTitle}
					class="input-bordered input w-full"
					placeholder="Book title"
				/>
			</div>
			<div class="form-control w-full">
				<label class="label"><span class="label-text">Author</span></label>
				<input
					type="text"
					bind:value={manualAuthor}
					class="input-bordered input w-full"
					placeholder="Author name (optional)"
				/>
			</div>
		</div>
	{:else}
		<!-- Search mode -->
		<div class="form-control relative w-full">
			<label class="label">
				<span class="label-text">Search Book</span>
			</label>
			<input
				type="text"
				bind:value={searchTerm}
				on:input={handleSearch}
				class="input-bordered input w-full"
				placeholder="Start typing a book title or author..."
			/>

			{#if loading}
				<div class="absolute z-50 mt-1 w-full rounded border bg-white pl-6 text-gray-500 shadow-lg">
					<span class="loading loading-ring loading-md"></span>
				</div>
			{:else if searchResults.length > 0}
				<ul
					class="absolute z-50 mt-1 max-h-40 w-full overflow-auto rounded border bg-white shadow-lg"
					style="top:100%; left:0;"
				>
					{#each searchResults as book}
						<li
							class="flex cursor-pointer justify-between p-2 hover:bg-gray-200"
							on:click={() => selectBook(book)}
						>
							<span>{book.work_title}</span>
							<span class="text-sm text-gray-500">by {book.author_name}</span>
						</li>
					{/each}
				</ul>
			{:else if searchPerformed && searchTerm.trim() && !selectedBook}
				<div class="mt-2 rounded border bg-amber-50 p-3">
					<p class="text-sm text-amber-800">No results found for "<strong>{searchTerm}</strong>".</p>
					<button class="btn btn-sm btn-outline btn-accent mt-2" on:click={enterManually}>
						Enter manually
					</button>
				</div>
			{/if}
		</div>

		<!-- Always-visible manual entry link -->
		{#if !selectedBook && searchTerm.trim()}
			<button class="btn btn-sm btn-ghost" on:click={enterManually}>
				Can't find your book? Enter it manually
			</button>
		{/if}
	{/if}

	<!-- Show selected book -->
	{#if !manualEntry && selectedBook}
		<div class="rounded border bg-green-50 p-3">
			<p class="text-sm text-green-800">Selected: <strong>{selectedBook}</strong></p>
			<button class="btn btn-xs btn-ghost mt-1" on:click={() => { selectedBook = ''; searchTerm = ''; }}>
				Clear
			</button>
		</div>
	{/if}

	<!-- Start and Finish dates -->
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
		<div class="form-control">
			<label class="label"><span class="label-text">Started Reading</span>
			<input type="date" bind:value={startDate} class="input-bordered input w-full" /></label>
		</div>
		<div class="form-control">
			<label class="label"><span class="label-text">Finished Reading</span>
			<input type="date" bind:value={finishDate} class="input-bordered input w-full" /></label>
		</div>
	</div>

	<!-- Number of pages -->
	<div class="form-control w-full">
		<label class="label"><span class="label-text">Number of Pages</span>
		<input
			type="text"
			min="1"
			bind:value={pages}
			inputmode="numeric"
			pattern="[0-9]*"
			class="input-bordered input w-full"
			placeholder="e.g., 350"
		/></label>
	</div>

	<!-- Rating -->
	<div class="form-control w-full">
		<label class="label"><span class="label-text">Your Rating: </span>
		<div class="rating gap-1">
			<input type="radio" name="rating" value={1} bind:group={rating} class="mask bg-amber-100 mask-heart" aria-label="1" />
			<input type="radio" name="rating" value={2} bind:group={rating} class="mask bg-rose-300 mask-heart" aria-label="2" />
			<input type="radio" name="rating" value={3} bind:group={rating} class="mask bg-teal-100 mask-heart" aria-label="3" />
			<input type="radio" name="rating" value={4} bind:group={rating} class="mask bg-pink-300 mask-heart" aria-label="4" />
			<input type="radio" name="rating" value={5} bind:group={rating} class="mask bg-purple-300 mask-heart" aria-label="5" />
			<input type="radio" name="rating" value={6} bind:group={rating} class="mask bg-blue-200 mask-heart" aria-label="6" />
			<input type="radio" name="rating" value={7} bind:group={rating} class="mask bg-red-200 mask-heart" aria-label="7" />
			<input type="radio" name="rating" value={8} bind:group={rating} class="mask bg-emerald-100 mask-heart" aria-label="8" />
			<input type="radio" name="rating" value={9} bind:group={rating} class="mask bg-pink-200 mask-heart" aria-label="9" />
			<input type="radio" name="rating" value={10} bind:group={rating} class="mask bg-violet-200 mask-heart" aria-label="10" />
		</div></label>
	</div>

	<!-- Thoughts -->
	<div class="form-control w-full">
		<label class="label"><span class="label-text">Your Thoughts:</span></label>
		<textarea
			bind:value={thoughts}
			class="textarea-bordered textarea h-40 w-full"
			placeholder="What did you think of the book?"
		></textarea>
	</div>

	<button class="btn btn-wide" on:click={submit}>Save Book Entry</button>
</div>