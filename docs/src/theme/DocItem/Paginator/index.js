import React from 'react';
import {useDoc} from '@docusaurus/theme-common/internal';
import DocPaginator from '@theme/DocPaginator';
import Link from '@docusaurus/Link';
/**
 * This extra component is needed, because <DocPaginator> should remain generic.
 * DocPaginator is used in non-docs contexts too: generated-index pages...
 */
export default function DocItemPaginator() {
  const {metadata} = useDoc();

  return (
    <dig>
      {(false) && (
      // {(metadata?.previous || metadata?.next) && (
        <div>
          <hr class="solid" />
          <h2>
            See also
          </h2>
          <ul>
          {metadata.previous && (
              <li>
                Prev page: <Link href={metadata.previous.permalink}>{metadata.previous.title}</Link>
              </li>
          )}
          {metadata.next && (
              <li>
                Next page: <Link href={metadata.next.permalink}>{metadata.next.title}</Link>
              </li>
          )}
          </ul>
        </div>
      )}
    </dig>
  );
}
