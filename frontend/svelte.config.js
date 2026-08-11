import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  kit: {
	version: {
		name: process.env.SVELTEKIT_APP_VERSION ?? 'betrayal-web-v1'
	},
    adapter: adapter({
      pages: '../internal/web/ui/dist',
      assets: '../internal/web/ui/dist',
      fallback: '200.html'
    })
  }
};

export default config;
