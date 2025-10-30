<script>
  import { onMount } from 'svelte';
  import { listUsers } from '$lib/api';
  import { goto } from '$app/navigation';

  let users = [];
  let loading = false;
  let error = null;
  let selected = null;

  onMount(async () => {
    loading = true;
    error = null;
    try {
      users = await listUsers();
    } catch (err) {
      error = err?.message || String(err);
    } finally {
      loading = false;
    }
  });

  function login() {
    if (!selected) {
      error = 'Please select a user to continue';
      return;
    }
    // store simple current user id in localStorage
    try {
      localStorage.setItem('currentUserId', String(selected));
    } catch (err) {
      console.warn('failed to persist user', err);
    }
    // navigate to home
    goto('/');
  }
</script>

<main class="p-4">
  <h1 class="text-2xl mb-4">Login / Select user</h1>

  {#if loading}
    <div>Loading users…</div>
  {:else}
    {#if error}
      <div class="text-red-600 mb-4">{error}</div>
    {/if}

    {#if users && users.length > 0}
      <form on:submit|preventDefault={login} class="space-y-3 max-w-md">
        {#each users as u}
          <label class="flex items-center gap-2">
            <input type="radio" name="user" bind:group={selected} value={u.id} />
            <span>{u.name || u.email} <small class="text-gray-500">#{u.id}</small></span>
          </label>
        {/each}

        <div class="pt-3">
          <button class="px-4 py-2 bg-blue-600 text-white" type="submit">Login</button>
          <a class="ml-3 text-blue-600 underline" href="/create_user">Create user</a>
        </div>
      </form>
    {:else}
      <div>No users found. <a class="text-blue-600 underline" href="/create_user">Create one</a>.</div>
    {/if}
  {/if}
</main>
