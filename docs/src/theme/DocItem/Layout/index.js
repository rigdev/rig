import React from 'react';
import clsx from 'clsx';
import {useWindowSize} from '@docusaurus/theme-common';
import {useDoc} from '@docusaurus/theme-common/internal';
import DocItemPaginator from '@theme/DocItem/Paginator';
import DocVersionBanner from '@theme/DocVersionBanner';
import DocVersionBadge from '@theme/DocVersionBadge';
import DocItemFooter from '@theme/DocItem/Footer';
import DocItemTOCMobile from '@theme/DocItem/TOC/Mobile';
import DocItemTOCDesktop from '@theme/DocItem/TOC/Desktop';
import DocItemContent from '@theme/DocItem/Content';
import styles from './styles.module.css';
import Feedback from '../../../components/Feedback';
/**
 * Decide if the toc should be rendered, on mobile or desktop viewports
 */
function useDocTOC() {
  const {frontMatter, toc} = useDoc();
  
  const windowSize = useWindowSize();
  const hidden = frontMatter.hide_table_of_contents;
  const hideFeedback = true;
  const hideFooter = frontMatter.hide_footer;
  const canRender = !hidden && toc.length > 0;
  const mobile = canRender ? <DocItemTOCMobile /> : undefined;
  const desktop =
    canRender && (windowSize === 'desktop' || windowSize === 'ssr') ? (
      <DocItemTOCDesktop />
    ) : undefined;
  return {
    hidden,
    mobile,
    desktop,
    hideFeedback,
    hideFooter
    
  };
}
export default function DocItemLayout({children}) {
  const docTOC = useDocTOC();
  return (
    <div className="row">
      <div className={clsx('col', !docTOC.hidden && styles.docItemCol)}>
        <DocVersionBanner />
        <div className={styles.docItemContainer}>
          <article>
            <DocVersionBadge />
            {docTOC.mobile}
            <DocItemContent>
                {children}
              <DocItemPaginator />
              {!docTOC.hideFeedback && (
                <Feedback 
                  event="survey_docs_helpful"
                  positiveQuestion="Is there anything that should improved?"
                  negativeQuestion="Please describe the issue you faced."
              />
              )}
            </DocItemContent>
            {!docTOC.hideFooter && (
              <DocItemFooter />
            )}
          </article>
        </div>
      </div>
      {docTOC.desktop && <div className="col col--3">{docTOC.desktop}</div>}
    </div>
  );
}
