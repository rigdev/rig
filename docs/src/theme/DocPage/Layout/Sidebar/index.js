import React, {useState, useCallback} from 'react';
import clsx from 'clsx';
import {prefersReducedMotion, ThemeClassNames} from '@docusaurus/theme-common';
import {useDocsSidebar} from '@docusaurus/theme-common/internal';
import {useLocation} from '@docusaurus/router';
import DocSidebar from '@theme/DocSidebar';
import ExpandButton from '@theme/DocPage/Layout/Sidebar/ExpandButton';
import styles from './styles.module.css';
// Reset sidebar state when sidebar changes
// Use React key to unmount/remount the children
// See https://github.com/facebook/docusaurus/issues/3414
function ResetOnSidebarChange({children}) {
  const sidebar = useDocsSidebar();
  return (
    <React.Fragment key={sidebar?.name ?? 'noSidebar'}>
      {children}
    </React.Fragment>
  );
}
export default function DocPageLayoutSidebar({
  sidebar,
  hiddenSidebarContainer,
  setHiddenSidebarContainer,
}) {
  const {pathname} = useLocation();
  const [hiddenSidebar, setHiddenSidebar] = useState(false);
  
  const [translateX, setTranslateX] = useState("translateX(0px)");
  const [currentPathname, setCurrentPathname] = useState(pathname);

  const animationDuration = 140;

  if(translateX === "translateX(-500px)"){
    setTimeout(function(){
      setTranslateX("translateX(0px)")
    }, animationDuration);
  }

  const toggleSidebar = useCallback(() => {
    if (hiddenSidebar) {
      setHiddenSidebar(false);
    }
    // onTransitionEnd won't fire when sidebar animation is disabled
    // fixes https://github.com/facebook/docusaurus/issues/8918
    if (!hiddenSidebar && prefersReducedMotion()) {
      setHiddenSidebar(true);
    }
    setHiddenSidebarContainer((value) => !value);
  }, [setHiddenSidebarContainer, hiddenSidebar]);

  const currentPathnames = currentPathname.split("/");
  const pathnames = pathname.split("/");
  let shouldAnimate = currentPathnames.length > 1 && pathnames.length > 1 && currentPathnames[1] !== pathnames[1];
  shouldAnimate = shouldAnimate && currentPathnames.length !== pathnames.length;

  if(shouldAnimate){
    //setTranslateX("translateX(-500px)");
    setTranslateX("translateX(0px)");
    setCurrentPathname(pathname);
  }

  return (
    <aside
      className={clsx(
        ThemeClassNames.docs.docSidebarContainer,
        styles.docSidebarContainer,
        hiddenSidebarContainer && styles.docSidebarContainerHidden,
      )}
      onTransitionEnd={(e) => {
        if (!e.currentTarget.classList.contains(styles.docSidebarContainer)) {
          return;
        }
        if (hiddenSidebarContainer) {
          setHiddenSidebar(true);
        }
      }}>
      <ResetOnSidebarChange>
        <div
          style={{
            transform: translateX,
            transition: `all ${animationDuration}ms`,
            marginTop: "15px",
          }}
          className={clsx(
            styles.sidebarViewport,
            hiddenSidebar && styles.sidebarViewportHidden,
          )}>
          <DocSidebar
            sidebar={sidebar}
            path={pathname}
            onCollapse={toggleSidebar}
            isHidden={hiddenSidebar}
          />
          {hiddenSidebar && <ExpandButton toggleSidebar={toggleSidebar} />}
        </div>
      </ResetOnSidebarChange>
    </aside>
  );
}