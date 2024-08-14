import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import {
  useDocById,
  findFirstSidebarItemLink,
} from '@docusaurus/plugin-content-docs/client';
import {usePluralForm} from '@docusaurus/theme-common';
import isInternalUrl from '@docusaurus/isInternalUrl';
import {translate} from '@docusaurus/Translate';
import Heading from '@theme/Heading';
import styles from './styles.module.css';
import IconBox from "../../components/IconBox";


function useCategoryItemsPlural() {
  const {selectMessage} = usePluralForm();
  return (count) =>
    selectMessage(
      count,
      translate(
        {
          message: '1 item|{count} items',
          id: 'theme.docs.DocCard.categoryDescription.plurals',
          description:
            'The default description for a category card in the generated index about how many items this category includes',
        },
        {count},
      ),
    );
}
function CardContainer({href, children, cover, color}) {
  return (
    <Link
      style={{
        backgroundImage: `url(${cover})`,
        backgroundSize: "contain",
        backgroundRepeat: "no-repeat",
        backgroundColor: color,
      }}
      href={href}
      className={clsx('card padding--md', styles.cardContainer)}>
      {children}
    </Link>
  );
}
function CardLayout({ href, icon, cover, title, description, image, color }) {
// function CardLayout({href, icon, title, description}) {
  return (
    <CardContainer href={href} cover={cover} color={color}>
      {image}
      <div className={clsx(styles.cardIcon)}>{icon}</div>
      <span className={clsx("text--truncate", styles.cardTitle)} title={title}>
        {title}
      </span>
      {description && (
        <p
          className={clsx('text--wrap', styles.cardDescription)}
          title={description}
        >
          {description}
        </p>
      )}
    </CardContainer>
  );
}
function CardCategory({item}) {
  const href = findFirstSidebarItemLink(item);
  const categoryItemsPlural = useCategoryItemsPlural();
  // Unexpected: categories that don't have a link have been filtered upfront
  if (!href) {
    return null;
  }
  return (
    <CardLayout
      href={href}
      icon="🗃️"
      title={item.label}
      description={item.description ?? categoryItemsPlural(item.items.length)}
    />
  );
}
function CardLink({item}) {
  // const icon = isInternalUrl(item.href) ? '📄️' : '🔗';
  const doc = useDocById(item.docId ?? undefined);
  let icon = "";
  if (item.customProps?.icon) {
    icon = <IconBox logo={item.customProps?.icon} />;
  } else {
    icon = isInternalUrl(item.href) ? "📄️" : "🔗";
  }

  return (
    <CardLayout
      href={item.href}
      icon={icon}
      title={item.label}
      cover={item.customProps?.cover}
      description={item.description ?? doc?.description}
    />
  );
}

function CardBlock({ item }) {
  const doc = useDocById(item.docId ?? undefined);
  let icon = "";
  return (
    <CardLayout
      href={item.href}
      title={item.label}
      description={item.description ?? doc?.description}
      color="white"
      cover={item.cover}
      image=<div
        style={{
          height: "180px",
        }}
      />
    />
  );
}

export default function DocCard({item}) {
  switch (item.type) {
    case 'link':
      return <CardLink item={item} />;
    case 'category':
      return <CardCategory item={item} />;
    case "block":
      return <CardBlock item={item} />;
    default:
      throw new Error(`unknown item type ${JSON.stringify(item)}`);
  }
}
