// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeGalaxy from 'starlight-theme-galaxy'
import mermaid from 'astro-mermaid';

// https://astro.build/config
export default defineConfig({
  site: 'https://boiler.iamabhinav.dev',
  markdown: {
    smartypants: false,
  },
  integrations: [
    mermaid({
      theme: 'forest',
      autoTheme: true
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
                { label: 'Supported Platforms', link: '/guides/remote-fetching#supported-platforms' },
                { label: 'Formats', link: '/guides/remote-fetching#understanding-formats' },
                { label: 'Registry Setup', link: '/guides/remote-fetching#setting-up-your-own-registry' },
              ],
            },
          ],
        },
        {
          label: 'Commands',
          autogenerate: { directory: 'commands' },
          collapsed: false
        },

      ],
    }),
  ],
});
