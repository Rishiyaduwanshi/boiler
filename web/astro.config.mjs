// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeGalaxy from 'starlight-theme-galaxy'
import mermaid from 'astro-mermaid';

// https://astro.build/config
export default defineConfig({
	markdown: {
		smartypants: false,
	},
	integrations: [
		mermaid({
			theme: 'forest',
			autoTheme : true
		}),
		starlight({
			plugins: [starlightThemeGalaxy()],
			title: 'Boiler',
			description: 'CLI tool for managing code snippets and project stacks',
			logo: {
				src: './src/assets/logo.svg',
			},
			favicon: '/favicon.svg',
			customCss: ['./src/styles/custom.css'],
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/rishiyaduwanshi/boiler' }
			],
			components: {
				SocialIcons: './src/components/CustomHeader.astro',
			},
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'guides/introduction' },
						{ label: 'Installation', slug: 'guides/installation' },
						{ label: 'Quick Start', slug: 'guides/quickstart' },
						{ label: 'Use Cases', slug: 'guides/usecases', badge: { text: 'Popular', variant: 'success' } },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Boiler Syntax', slug: 'guides/syntax' },
						{
							label: 'Remote Fetching',
							badge: { text: 'New', variant: 'tip' },
							collapsed: false,
							items: [
								{ label: 'Overview', slug: 'guides/remote-fetching' },
								{ label: 'Formats', link: '/guides/remote-fetching#understanding-formats' },
								{ label: 'GitHub Fetch', link: '/guides/remote-fetching#fetching-from-github' },
								{ label: 'Registry Fetch', link: '/guides/remote-fetching#using-a-registry' },
								{ label: 'Custom Domain', link: '/guides/remote-fetching#fetching-from-custom-websites' },
								{ label: 'Direct URL', link: '/guides/remote-fetching#fetching-from-direct-urls' },
							],
						},
					],
				},
				{
					label: 'Commands',
					autogenerate: { directory: 'commands' },
					collapsed :  false
				},

			],
		}),
	],
});
