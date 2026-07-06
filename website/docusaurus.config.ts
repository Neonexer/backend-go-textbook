import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";
import { themes } from "prism-react-renderer";

const config: Config = {
  title: "Бэкенд на Go",
  tagline: "Цифровой учебник по бэкенд-разработке на Go — от первого сервера до YouTube-клона",
  favicon: "img/favicon.ico",

  url: "https://neonexer.github.io",
  baseUrl: "/backend-go-textbook/",

  organizationName: "Neonexer",
  projectName: "backend-go-textbook",

  onBrokenLinks: "throw",
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: "warn",
    },
  },

  i18n: {
    defaultLocale: "ru",
    locales: ["ru"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          sidebarPath: "./sidebars.ts",
          routeBasePath: "/",
          editUrl: "https://github.com/go-course/backend-go-textbook/tree/main/website/",
          showLastUpdateTime: false,
        },
        blog: false,
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  themes: [
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      {
        hashed: true,
        language: ["ru", "en"],
        indexDocs: true,
        indexBlog: false,
        docsRouteBasePath: "/",
      },
    ],
  ],

  themeConfig: {
    image: "img/social-card.png",
    colorMode: {
      defaultMode: "dark",
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: "Бэкенд на Go",
      logo: {
        alt: "Go Gopher",
        src: "img/logo.svg",
      },
      items: [
        {
          type: "docSidebar",
          sidebarId: "textbook",
          position: "left",
          label: "Содержание",
        },
        {
          href: "https://github.com/go-course/backend-go-textbook",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Учебник",
          items: [
            {
              label: "Введение",
              to: "/intro",
            },
            {
              label: "Проект 1: REST API",
              to: "/project-1-blog",
            },
            {
              label: "Проект 2: Маркетплейс",
              to: "/project-2-marketplace",
            },
            {
              label: "Проект 3: YouTube-клон",
              to: "/project-3-youtube-clone",
            },
          ],
        },
        {
          title: "Сообщество",
          items: [
            {
              label: "GitHub",
              href: "https://github.com/go-course/backend-go-textbook",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Бэкенд на Go. Построено с Docusaurus.`,
    },
    prism: {
      theme: themes.github,
      darkTheme: themes.dracula,
      additionalLanguages: ["go", "bash", "yaml", "docker", "protobuf", "sql"],
      // Docusaurus 3.x uses Prism via prism-react-renderer
    },
    docs: {
      sidebar: {
        hideable: true,
        autoCollapseCategories: false,
      },
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
