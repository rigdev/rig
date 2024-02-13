// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Rig Docs',
  tagline: 'Explore and learn how to use Rig',
  favicon: 'img/favicon.ico',
  // Set the production url of your site here
  url: 'https://rig.dev',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'Rigdev', // Usually your GitHub org/user name.
  projectName: 'docs', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [
    [
      '@docusaurus/plugin-google-tag-manager',
      {
        containerId: 'GTM-P2VBJ6K',
      }
    ],
    (context, opts) => {
      return {
        name: 'tidio',
        injectHtmlTags({content}) {
          return {
            postBodyTags: [`<script src="//code.tidio.co/nzmuyu07fgt8wm0jkj4dv0rj9ti2djfw.js" async></script>`],
          };
        }
      };
    }
  ],

  scripts: [
    "https://js-eu1.hs-scripts.com/139684927.js",
    {
      src: '/js/segment.js'
    },
  ],

  stylesheets: [
    "https://doc.gendocu.com/widget/documentation.css"
  ],
  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./menu.js'),
          // Please change this to your repo.
          editUrl: 'https://github.com/rigdev/rig/edit/main/docs/',
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
          routeBasePath: '/', // Serve the docs at the site's root
          // breadcrumbs: false,
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: 'img/docusaurus-social-card.jpg',
      docs: {
        sidebar: {
          autoCollapseCategories: true,
        }
      },
      colorMode: {
        defaultMode: "light",
        respectPrefersColorScheme: true,
      },
      navbar: {
        hideOnScroll: false,
        logo: {
          alt: 'Rig Logo',
          src: 'img/logo.png',
          srcDark: 'img/logo-dark.png',
          width: 100,
          style: {
            objectFit: "contain"
          }
        },
      },
      prism: {
        theme: require('prism-react-renderer/themes/vsDark'),
      },
    }),
};

module.exports = config;
