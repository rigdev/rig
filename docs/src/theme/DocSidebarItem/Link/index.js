import React from 'react';
import clsx from 'clsx';
import {ThemeClassNames} from '@docusaurus/theme-common';
import {isActiveSidebarItem} from '@docusaurus/theme-common/internal';
import Link from '@docusaurus/Link';
import isInternalUrl from '@docusaurus/isInternalUrl';
import IconExternalLink from '@theme/Icon/ExternalLink';
import styles from './styles.module.css';
import DynamicBiIcon from '../../../components/DynamicBiIcon';
import IconBox from '../../../components/IconBox';

export default function DocSidebarItemLink({
  item,
  onItemClick,
  activePath,
  level,
  index,
  ...props
}) {
  const {href, label, className, autoAddBaseUrl, customProps} = item;
  const isActive = isActiveSidebarItem(item, activePath);
  const isInternalLink = isInternalUrl(href);

  return (
    <li
      className={clsx(
        ThemeClassNames.docs.docSidebarItemLink,
        ThemeClassNames.docs.docSidebarItemLinkLevel(level),
        'menu__list-item',
        className,
      )}
      key={label}>
      <Link
        className={clsx(
          'menu__link',
          !isInternalLink && styles.menuExternalLink,
          {
            'menu__link--active': isActive && !customProps?.boxed,
          },
        )}
        autoAddBaseUrl={autoAddBaseUrl}
        aria-current={isActive ? 'page' : undefined}
        to={href}
        {...(isInternalLink && {
          onClick: onItemClick ? () => onItemClick(item) : undefined,
        })}
        {...props}>
          <div style={{marginRight: "5px", marginTop: "3px"}}>
          {!customProps?.boxed && customProps?.sidebar_icon && (
            <DynamicBiIcon name={customProps.sidebar_icon} />
          )}
          {customProps?.boxed && customProps?.sidebar_icon && (
            <IconBox logo={customProps.sidebar_icon} size={40}/>
          )}
         </div>
         {customProps?.boxed && (
          <div style={{fontSize: "16px", marginLeft: "10px", color: "var(--ifm-color-emphasis-900)"}}>
            {label}
          </div>
         )}
         {!customProps?.boxed && (
          label
         )}
        {!isInternalLink && <IconExternalLink />}
      </Link>
    </li>
  );
}