import React from 'react';
import clsx from 'clsx';
import {
  useCurrentSidebarCategory,
  filterDocCardListItems,
} from '@docusaurus/theme-common';
import DocCard from '@theme/DocCard';
function DocCardListForCurrentSidebarCategory({className}) {
  const category = useCurrentSidebarCategory();
  return <DocCardList items={category.items} className={className} />;
}
export default function DocCardList(props) {
  const {items, className} = props;
  if (!items) {
    return <DocCardListForCurrentSidebarCategory {...props} />;
  }
  const filteredItems = filterDocCardListItems(items);
  return (
    <section className={clsx('row', 'padding-left--sm', 'padding-right--sm', className)}>
      {filteredItems.map((item, index) => {
        return (
          <article key={index} className={`col ${item.customProps?.col ?? "col--6"} padding--sm`}>
            <DocCard item={item} />
          </article>
        )
      })}
    </section>
  );
}
