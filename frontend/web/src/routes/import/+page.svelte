<script lang="ts">
	import { onMount } from 'svelte';

	type ImportStatus = 'idle' | 'uploading' | 'success' | 'error';

	let authorsStatus: ImportStatus = 'idle';
	let worksStatus: ImportStatus = 'idle';
	let authorsMessage = '';
	let worksMessage = '';

	async function uploadFile(fileType: 'authors' | 'works') {
		const input = document.createElement('input');
		input.type = 'file';
		input.accept = '.txt,.json,.jsonl,.gz';
		input.onchange = async () => {
			const file = input.files?.[0];
			if (!file) return;

			const statusRef = fileType === 'authors' ? 'authors' : 'works';
			if (fileType === 'authors') {
				authorsStatus = 'uploading';
				authorsMessage = `Uploading ${file.name}...`;
			} else {
				worksStatus = 'uploading';
				worksMessage = `Uploading ${file.name}...`;
			}

			const formData = new FormData();
			formData.append('file', file);
			formData.append('type', fileType);

			try {
				const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
				const res = await fetch(`${apiUrl}/api/import`, {
					method: 'POST',
					body: formData
				});

				if (!res.ok) {
					const text = await res.text();
					if (fileType === 'authors') {
						authorsStatus = 'error';
						authorsMessage = `Failed: ${text}`;
					} else {
						worksStatus = 'error';
						worksMessage = `Failed: ${text}`;
					}
					return;
				}

				if (fileType === 'authors') {
					authorsStatus = 'success';
					authorsMessage = `Uploaded ${file.name} successfully`;
				} else {
					worksStatus = 'success';
					worksMessage = `Uploaded ${file.name} successfully`;
				}
			} catch (err) {
				if (fileType === 'authors') {
					authorsStatus = 'error';
					authorsMessage = `Error: ${err}`;
				} else {
					worksStatus = 'error';
					worksMessage = `Error: ${err}`;
				}
			}
		};
		input.click();
	}

	function statusClass(status: ImportStatus): string {
		switch (status) {
			case 'success': return 'text-success';
			case 'error': return 'text-error';
			case 'uploading': return 'text-info';
			default: return '';
		}
	}
</script>

<div class="mx-auto max-w-3xl space-y-8">
	<h1 class="text-3xl font-bold">Import</h1>
	<p class="opacity-70">
		Upload OpenLibrary dump files to populate the search database.
		Authors must be uploaded before works.
	</p>

	<div class="grid gap-6 md:grid-cols-2">
		<!-- Authors -->
		<div class="card bg-base-100 shadow-sm">
			<div class="card-body">
				<h2 class="card-title">Authors</h2>
				<p class="text-sm opacity-70">
					Upload the OpenLibrary authors dump file (tab-separated, JSON per line).
				</p>
				<div class="card-actions justify-end mt-4">
					<button
						class="btn btn-primary"
						class:loading={authorsStatus === 'uploading'}
						disabled={authorsStatus === 'uploading'}
						on:click={() => uploadFile('authors')}
					>
						{authorsStatus === 'uploading' ? 'Uploading...' : 'Choose File'}
					</button>
				</div>
				{#if authorsMessage}
					<p class="text-sm mt-2 {statusClass(authorsStatus)}">{authorsMessage}</p>
				{/if}
			</div>
		</div>

		<!-- Works -->
		<div class="card bg-base-100 shadow-sm">
			<div class="card-body">
				<h2 class="card-title">Works</h2>
				<p class="text-sm opacity-70">
					Upload the OpenLibrary works dump file (tab-separated, JSON per line).
				</p>
				<div class="card-actions justify-end mt-4">
					<button
						class="btn btn-primary"
						class:loading={worksStatus === 'uploading'}
						disabled={worksStatus === 'uploading'}
						on:click={() => uploadFile('works')}
					>
						{worksStatus === 'uploading' ? 'Uploading...' : 'Choose File'}
					</button>
				</div>
				{#if worksMessage}
					<p class="text-sm mt-2 {statusClass(worksStatus)}">{worksMessage}</p>
				{/if}
			</div>
		</div>
	</div>

	<div class="rounded border bg-base-200 p-4 text-sm opacity-70">
		<h3 class="font-semibold mb-1">File format</h3>
		<p>
			OpenLibrary dump files are tab-separated: each line has an ID, a tab, then a JSON object.
			Get them from
			<a href="https://openlibrary.org/developers/dumps" target="_blank" rel="noopener" class="link link-primary">
				openlibrary.org/developers/dumps
			</a>.
		</p>
	</div>
</div>