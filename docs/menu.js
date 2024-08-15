/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  // homepage: [{type: 'autogenerated', dirName: '.'}],

  // But you can create a sidebar manually
  // myAutogeneratedSidebar: [
  //   {
  //     type: 'autogenerated',
  //     dirName: '.', // '.' means the current docs folder
  //   },
  // ],

  homepage: [
    {
      type: "html",
      value: "Overview",
      className: "homepage-sidebar-divider",
    },

    {
      type: "category",
      label: "Getting Started",
      className: "homepage-sidebar-item",
      link: {
        type: "doc",
        id: "getting-started",
      },
      customProps: {
        sidebar_icon: "BiRocket",
      },
      collapsed: true,
      items: [
        {
          type: "doc",
          id: "getting-started/cli-install",
          label: "CLI Installation",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "getting-started/platform-installation",
          label: "Platform Installation",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "getting-started/first-capsule",
          label: "Your first Capsule",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "getting-started/next-steps",
          label: "Next Steps",
          className: "docpage-sidebar-item",
        },
      ],
    },
    {
      type: "html",
      value: "Platform",
      className: "homepage-sidebar-divider",
    },
    {
      type: "doc",
      label: "Capsules",
      id: "platform/capsules",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiCapsule",
      },
    },
    {
      type: "doc",
      id: "platform/images",
      label: "Images",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiBox",
      },
    },
    {
      type: "doc",
      id: "platform/rollouts",
      label: "Rollouts",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "TbPlayerTrackNextFilled",
      },
    },
    {
      type: "doc",
      id: "platform/config-files",
      label: "Config Files",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiSolidFile",
      },
    },
    {
      type: "doc",
      id: "platform/container-settings",
      label: "Container Settings",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "SiLinuxcontainers",
      },
    },
    {
      type: "doc",
      id: "platform/network-interfaces",
      label: "Network Interfaces",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiSolidNetworkChart",
      },
    },
    {
      type: "category",
      link: {
        type: "doc",
        id: "platform/scale",
      },
      label: "Scale",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiArea",
      },
      items: [
        {
          type: "doc",
          id: "platform/custom-metrics",
          label: "Custom Metrics",
          className: "docpage-sidebar-item",
          customProps: {
            sidebar_icon: "BiAbacus",
          },
        },
        {
          type: "doc",
          id: "platform/custom-metrics-example",
          label: "Custom Metrics - Example",
          className: "docpage-sidebar-item",
          customProps: {
            sidebar_icon: "BiHardHat",
          },
        },
      ],
    },
    {
      type: "doc",
      id: "platform/instance-overview",
      label: "Instance Overview",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiOutline",
      },
    },
    {
      type: "doc",
      id: "platform/cronjobs",
      label: "Cron Jobs",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiCalendar",
      },
    },
    {
      type: "doc",
      id: "platform/service-accounts",
      label: "Service Accounts",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiKey",
      },
    },
    {
      type: "doc",
      id: "platform/rbac",
      label: "RBAC",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiLock",
      },
    },
    {
      type: "html",
      value: "Operator Manual",
      className: "homepage-sidebar-divider",
    },
    {
      type: "category",
      label: "Setup Guide",
      link: { type: "doc", id: "operator-manual/setup-guide" },
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiWrench",
      },
      collapsed: true,
      items: [
        {
          type: "category",
          link: { type: "doc", id: "operator-manual/setup-guide/operator" },
          label: "Operator",
          className: "docpage-sidebar-item",
          customProps: {
            sidebar_icon: "BiChip",
          },
          items: [
            {
              type: "doc",
              id: "operator-manual/setup-guide/operator/configuration-secrets",
              label: "Configuration as Secrets",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              label: "Plugins",
              id: "operator-manual/setup-guide/operator/plugins",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              label: "Networking",
              id: "operator-manual/setup-guide/operator/networking",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/setup-guide/operator/autoscaler",
              label: "Autoscaler and Custom Metrics",
              className: "docpage-sidebar-item",
            },
          ],
        },
        {
          type: "category",
          label: "Platform",
          link: { type: "doc", id: "operator-manual/setup-guide/platform" },
          className: "docpage-sidebar-item",
          collapsed: true,
          customProps: {
            sidebar_icon: "BiLaptop",
          },
          items: [
            {
              type: "doc",
              label: "Multicluster",
              id: "operator-manual/setup-guide/platform/multicluster",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/setup-guide/platform/database",
              label: "Database",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/setup-guide/platform/network",
              label: "Networking",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/setup-guide/platform/sso",
              label: "Single Sign-on",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/setup-guide/platform/notifications",
              label: "Notifications",
              className: "docpage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/setup-guide/platform/container-registries",
              label: "Container Registries",
              className: "docpage-sidebar-item",
            },
          ],
        },
      ],
    },
    {
      type: "category",
      link: {
        type: "doc",
        id: "operator-manual/ci-cd",
      },
      label: "CI/CD",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "TbRepeat",
      },
      items: [
        {
          type: "doc",
          id: "operator-manual/ci-cd/github-actions",
          label: "GitHub Actions",
          className: "docpage-sidebar-item",
          customProps: {
            sidebar_icon: "TbBrandGithub",
          },
        },
        {
          type: "doc",
          id: "operator-manual/ci-cd/gitlab-pipelines",
          label: "GitLab CI/CD Pipelines",
          className: "docpage-sidebar-item",
          customProps: {
            sidebar_icon: "TbBrandGitlab",
          },
        },
      ],
    },
    {
      type: "doc",
      id: "operator-manual/architecture",
      label: "Architecture",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiSolidLayout",
      },
    },
    {
      type: "doc",
      id: "operator-manual/environments",
      label: "Environments",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiSolidServer",
      },
    },
    {
      type: "doc",
      id: "operator-manual/onboarding",
      label: "Live Onboarding",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiTask",
      },
    },
    {
      type: "doc",
      id: "operator-manual/gitops",
      label: "GitOps",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "SiGit",
      },
    },
    {
      type: "doc",
      id: "operator-manual/review",
      label: "Review Process",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiCheckSquare",
      },
    },
    //collapsed: true,
    //items: [
    //  {
    //    type: "doc",
    //    id: "operator-manual/gitops/setup-with-flux",
    //    label: "Setup with Flux",
    //    className: "docpage-sidebar-item",
    //    customProps: {
    //      sidebar_icon: "SiFlux",
    //    },
    //  },
    //  {
    //    type: "doc",
    //    id: "operator-manual/gitops/setup-with-argo-cd",
    //    label: "Setup with Argo CD",
    //    className: "docpage-sidebar-item",
    //    customProps: {
    //      sidebar_icon: "SiArgo",
    //    },
    //  },
    //],
    // },
    // {
    //   type: "doc",
    //   id: "operator-manual/secrets",
    //   label: "Secrets",
    //   className: "homepage-sidebar-item",
    //   customProps: {
    //     sidebar_icon: "BiKey",
    //   },
    // },
    {
      type: "doc",
      id: "operator-manual/crd-operator",
      label: "CRD & Operator",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiDna",
      },
    },
    {
      type: "doc",
      id: "operator-manual/cli",
      label: "Rig Ops CLI",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiSolidTerminal",
      },
    },
    {
      type: "category",
      label: "Plugins",
      link: { type: "doc", id: "operator-manual/plugins/plugins" },
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiInjection",
      },
      collapsed: true,
      items: [
        // {
        //   type: "doc",
        //   label: "Writing third-party plugins",
        //   id: "operator-manual/plugins/thirdparty",
        //   className: "homepage-sidebar-item",
        //   customProps: {
        //     sidebar_icon: "BiPencil",
        //   },
        // },
        {
          type: "category",
          label: "Capsule Steps",
          link: {type: "doc", id: "operator-manual/plugins/capsulesteps"},
          className: "homepage-sidebar-item",
          collapsed: true,
          customProps: {
            sidebar_icon: "BiCapsule",
          },
          items: [
            {
              type: "doc",
              id: "operator-manual/plugins/capsulesteps/service_account",
              label: "Service Account",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/capsulesteps/deployment",
              label: "Deployment",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/capsulesteps/ingress_routes",
              label: "Ingress Routes",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/capsulesteps/cron_jobs",
              label: "Cron Jobs",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/capsulesteps/service_monitor",
              label: "Service Monitor",
              className: "homepage-sidebar-item",
            },
          ],
        },
        {
          type: "category",
          label: "Builtin",
          link: { type: "doc", id: "operator-manual/plugins/builtin" },
          className: "homepage-sidebar-item",
          collapsed: true,
          customProps: {
            sidebar_icon: "BiChip",
          },
          items: [
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/annotations",
              label: "Annotations",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/datadog",
              label: "Datadog",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/env_mapping",
              label: "Env Mapping",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/google_cloud_sql_auth_proxy",
              label: "Google Cloud SQL Auth Proxy",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/init_container",
              label: "Init Container",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/object_template",
              label: "Object Template",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/placement",
              label: "Placement",
              className: "homepage-sidebar-item",
            },
            {
              type: "doc",
              id: "operator-manual/plugins/builtin/sidecar",
              label: "Sidecar",
              className: "homepage-sidebar-item",
            },
          ],
        },
        {
          type: "doc",
          id: "operator-manual/plugins/thirdparty",
          label: "Custom plugins",
          className: "homepage-sidebar-item",
          customProps: {
            sidebar_icon: "BiPencil",
          },
        },
        {
          type: "doc",
          id: "operator-manual/plugins/examples",
          label: "Examples",
          className: "homepage-sidebar-item",
          customProps: {
            sidebar_icon: "BiBookmarks",
          },
        }
      ],
    },
    /*
    {
      type: "html",
      value: "Cloud Providers",
      className: "homepage-sidebar-divider",
    },
    {
      type: "doc",
      id: "additional-links",
      label: "Google Kubernetes Engine",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiLink",
      },
    },
    {
      type: "doc",
      id: "additional-links",
      label: "AWS EKS",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiLink",
      },
    },
*/
    {
      type: "html",
      value: "Additional Resources",
      className: "homepage-sidebar-divider",
    },
    {
      type: "category",
      label: "Reference Documentation",
      link: { type: "doc", id: "api" },
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiSolidFile",
      },
      collapsed: true,
      items: [
        {
          type: "doc",
          id: "api/platform-api",
          label: "Platform API Reference",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "api/platformv1",
          label: "platform.rig.dev/v1",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "api/config/v1alpha1",
          label: "config.rig.dev/v1alpha1",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "api/v1alpha1",
          label: "rig.dev/v1alpha1",
          className: "docpage-sidebar-item",
        },
        {
          type: "doc",
          id: "api/v1alpha2",
          label: "rig.dev/v1alpha2",
          className: "docpage-sidebar-item",
        },
      ],
    },

    {
      type: "doc",
      id: "additional-links",
      label: "Additional Links",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiLink",
      },
    },
    {
      type: "doc",
      id: "usage",
      label: "Usage",
      className: "homepage-sidebar-item",
      customProps: {
        sidebar_icon: "BiKey",
      },
    },
  ],
};

module.exports = sidebars;
