import React from 'react';
import {ThemeClassNames} from '@docusaurus/theme-common';
import Button from '../../components/Button';
export default function EditThisPage({editUrl}) {
  return (
    <a
      href={editUrl}
      target="_blank"
      rel="noreferrer noopener"
      className={ThemeClassNames.common.editThisPage}>
      <Button
        id="theme.common.editThisPage"
        description="The link label to edit the current page">
        Edit this page
      </Button>
    </a>
  );
}
