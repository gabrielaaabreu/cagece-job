<script>
	import { createUser } from "$lib/api";
	let name = "";
	let email = "";
	let message = null;
	let loading = false;

	async function onSubmit(e) {
		e.preventDefault();
		message = null;
		loading = true;
		try {
			const res = await createUser({ name, email });
			if (res && res.error) {
				message = { type: 'error', text: res.error };
			} else {
				message = { type: 'success', text: `User created with id ${res.id}` };
				name = '';
				email = '';
			}
		} catch (err) {
			message = { type: 'error', text: err.message };
		} finally {
			loading = false;
		}
	}
</script>

<main class="p-4">
	<h1 class="text-2xl mb-4">Create user</h1>
	{#if message}
		<div class="mb-4 {message.type === 'error' ? 'text-red-600' : 'text-green-600'}">{message.text}</div>
	{/if}
	<form on:submit|preventDefault={onSubmit} class="space-y-4 max-w-md">
		<div>
			<label class="block mb-1">Name</label>
			<input class="w-full border p-2" bind:value={name} required />
		</div>
		<div>
			<label class="block mb-1">Email</label>
			<input type="email" class="w-full border p-2" bind:value={email} required />
		</div>
		<div>
			<button class="px-4 py-2 bg-blue-600 text-white" type="submit" disabled={loading}>
				{#if loading}Creating...{:else}Create{/if}
			</button>
		</div>
	</form>
</main>
