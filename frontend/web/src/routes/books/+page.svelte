<script lang="ts">
  import { onMount } from 'svelte';

  type ApiBook = {
    ID: string;
    Book: string;
    Rating: number;
    StartDate: { Time: string; Valid: boolean };
    FinishDate: { Time: string; Valid: boolean };
    Pages: { Int32: number; Valid: boolean };
  };

  type Book = {
    id: string;
    book: string;
    rating: number;
    startDate: string | null;
    finishDate: string | null;
    pages: number | null;
  };

  function formatDate(iso: string | null): string {
    if (!iso) return '-';
    const d = new Date(iso);
    const day = String(d.getDate()).padStart(2, '0');
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const year = d.getFullYear();
    return `${day}-${month}-${year}`;
  }

  type SortKey = 'book' | 'rating' | 'startDate' | 'finishDate' | 'pages';
  let sortKey: SortKey = 'book';
  let sortAsc = true;

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      sortAsc = !sortAsc;
    } else {
      sortKey = key;
      sortAsc = true;
    }
  }

  $: sortedRows = [...rows].sort((a, b) => {
    let va: string | number = '';
    let vb: string | number = '';

    switch (sortKey) {
      case 'book':
        va = a.book.toLowerCase();
        vb = b.book.toLowerCase();
        break;
      case 'rating':
        va = a.rating;
        vb = b.rating;
        break;
      case 'startDate':
        va = a.startDate ?? '';
        vb = b.startDate ?? '';
        break;
      case 'finishDate':
        va = a.finishDate ?? '';
        vb = b.finishDate ?? '';
        break;
      case 'pages':
        va = a.pages ?? 0;
        vb = b.pages ?? 0;
        break;
    }

    if (va < vb) return sortAsc ? -1 : 1;
    if (va > vb) return sortAsc ? 1 : -1;
    return 0;
  });

  let rows: Book[] = [];

  onMount(async () => {
    const apiUrl = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
    const res = await fetch(`${apiUrl}/api/books`);
    const data: ApiBook[] = await res.json();

    rows = data.map((b) => ({
      id: b.ID,
      book: b.Book,
      rating: b.Rating,
      startDate: b.StartDate?.Valid ? b.StartDate.Time : null,
      finishDate: b.FinishDate?.Valid ? b.FinishDate.Time : null,
      pages: b.Pages?.Valid ? b.Pages.Int32 : null
    }));
  });
</script>

<div class="overflow-x-auto">
  <table class="table table-xs">
    <thead>
      <tr>
        <th>#</th>
        <th>
          <button class="btn btn-ghost btn-xs" on:click={() => toggleSort('book')}>
            Book
            {#if sortKey === 'book'}<span>{sortAsc ? '↑' : '↓'}</span>{/if}
          </button>
        </th>
        <th>
          <button class="btn btn-ghost btn-xs" on:click={() => toggleSort('rating')}>
            Rating
            {#if sortKey === 'rating'}<span>{sortAsc ? '↑' : '↓'}</span>{/if}
          </button>
        </th>
        <th>
          <button class="btn btn-ghost btn-xs" on:click={() => toggleSort('startDate')}>
            Start Date
            {#if sortKey === 'startDate'}<span>{sortAsc ? '↑' : '↓'}</span>{/if}
          </button>
        </th>
        <th>
          <button class="btn btn-ghost btn-xs" on:click={() => toggleSort('finishDate')}>
            Finish Date
            {#if sortKey === 'finishDate'}<span>{sortAsc ? '↑' : '↓'}</span>{/if}
          </button>
        </th>
        <th>
          <button class="btn btn-ghost btn-xs" on:click={() => toggleSort('pages')}>
            Pages
            {#if sortKey === 'pages'}<span>{sortAsc ? '↑' : '↓'}</span>{/if}
          </button>
        </th>
      </tr>
    </thead>

    <tbody>
      {#each sortedRows as entry, i}
        <tr>
          <th>{i + 1}</th>
          <td>{entry.book}</td>
          <td>{entry.rating}</td>
          <td>{formatDate(entry.startDate)}</td>
          <td>{formatDate(entry.finishDate)}</td>
          <td>{entry.pages ?? '-'}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>