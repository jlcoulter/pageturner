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
        <th>Book</th>
        <th>Rating</th>
        <th>Start Date</th>
        <th>Finish Date</th>
        <th>Pages</th>
      </tr>
    </thead>

    <tbody>
      {#each rows as entry, i}
        <tr>
          <th>{i + 1}</th>
          <td>{entry.book}</td>
          <td>{entry.rating}</td>
          <td>{entry.startDate ?? '-'}</td>
          <td>{entry.finishDate ?? '-'}</td>
          <td>{entry.pages ?? '-'}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
