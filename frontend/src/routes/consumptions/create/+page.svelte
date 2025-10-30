<script>
  import { onMount } from 'svelte';
  import { listUsers, createConsumption } from '$lib/api';

  let users = [];
  let selectedUser = null;
  let year = new Date().getFullYear();
  let month = new Date().getMonth() + 1;
  let cubic_meters = '';
  let loading = false;
  let message = null;

  onMount(async () => {
    loading = true;
    message = null;
    try {
      users = await listUsers();
      const cur = localStorage.getItem('currentUserId');
      if (cur) selectedUser = Number(cur);
    } catch (err) {
      message = { type: 'error', text: err?.message || String(err) };
    } finally {
      loading = false;
    }
  });

  async function onSubmit(e) {
    e.preventDefault();
    message = null;
    if (!selectedUser) {
      message = { type: 'error', text: 'Please select a user.' };
      return;
    }
    if (!year || !month || cubic_meters === '') {
      message = { type: 'error', text: 'Please fill all fields.' };
      return;
    }
    loading = true;
    try {
      const payload = {
        year: Number(year),
        month: Number(month),
        cubic_meters: Number(cubic_meters),
      };
      const res = await createConsumption(selectedUser, payload);
      if (res && res.error) {
        message = { type: 'error', text: res.error };
      } else {
        message = { type: 'success', text: `Consumption created (id ${res.id})` };
        cubic_meters = '';
      }
    } catch (err) {
      message = { type: 'error', text: err?.message || String(err) };
    } finally {
      loading = false;
    }
  }
</script>

<main class="p-4">
  <h1 class="text-2xl mb-4">Submit monthly consumption</h1>

  {#if message}
    <div class="mb-4 {message.type === 'error' ? 'text-red-600' : 'text-green-600'}">{message.text}</div>
  {/if}

  <form on:submit|preventDefault={onSubmit} class="space-y-4 max-w-md">
    <div>
      <label class="block mb-1">User</label>
      {#if users && users.length > 0}
        <select bind:value={selectedUser} class="w-full border p-2">
          <option value="" disabled selected={!selectedUser}>-- choose user --</option>
          {#each users as u}
            <option value={u.id}>{u.name ? `${u.name} (${u.email})` : u.email} — #{u.id}</option>
          {/each}
        </select>
      {:else}
        <div>No users found. Create one at <a href="/create_user" class="text-blue-600 underline">/create_user</a></div>
      {/if}
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="block mb-1">Year</label>
        <input type="number" bind:value={year} min="2000" class="w-full border p-2" />
      </div>
      <div>
        <label class="block mb-1">Month</label>
        <select bind:value={month} class="w-full border p-2">
          {#each Array(12) as _, i}
            <option value={i + 1}>{i + 1}</option>
          {/each}
        </select>
      </div>
    </div>

    <div>
      <label class="block mb-1">Cubic meters</label>
      <input type="number" step="0.01" bind:value={cubic_meters} class="w-full border p-2" required />
    </div>

    <div>
      <button class="px-4 py-2 bg-blue-600 text-white" type="submit" disabled={loading}>
        {#if loading}Submitting...{:else}Submit{/if}
      </button>
    </div>
  </form>
</main>
