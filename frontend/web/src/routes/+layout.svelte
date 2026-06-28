<script lang="ts">
	import './layout.css';

	let { children } = $props();
</script>

<div class="flex min-h-screen flex-col">
	<!-- Navbar -->
	<nav class="navbar bg-base-100 px-4 py-3 shadow-sm">
		<a class="btn text-xl btn-ghost" href="/">Home</a>
		<a class="btn text-xl btn-ghost" href="/books">Books</a>
		<a class="btn text-xl btn-ghost" href="/commonplace">Commonplace</a>
		<a class="btn text-xl btn-ghost" href="/import">Import</a>
	</nav>

	<!-- Main content -->
	<main class="flex-1 px-6 py-8">
		{@render children()}
	</main>

	<!-- Footer -->
	<footer class="footer-center footer bg-base-300 p-4 text-base-content sm:footer-horizontal">
		<p>Copyright © {new Date().getFullYear()} Bookplace</p>
	</footer>
</div>

<style>
  :global(body) {
    font-family: 'Times New Roman', serif;
  }
</style>
