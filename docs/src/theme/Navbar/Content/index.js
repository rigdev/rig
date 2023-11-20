import React from 'react';
import {useThemeConfig, ErrorCauseBoundary} from '@docusaurus/theme-common';
import {
  splitNavbarItems,
  useNavbarMobileSidebar,
} from '@docusaurus/theme-common/internal';
import NavbarItem from '@theme/NavbarItem';
import NavbarColorModeToggle from '@theme/Navbar/ColorModeToggle';
import SearchBar from '@theme/SearchBar';
import NavbarMobileSidebarToggle from '@theme/Navbar/MobileSidebar/Toggle';
import NavbarLogo from '@theme/Navbar/Logo';
import NavbarSearch from '@theme/Navbar/Search';
import styles from './styles.module.css';
import Button from '../../../components/Button';
import DynamicBiIcon from '../../../components/DynamicBiIcon';
import Link from '@docusaurus/Link';
function useNavbarItems() {
  // TODO temporary casting until ThemeConfig type is improved
  return useThemeConfig().navbar.items;
}
function NavbarItems({items}) {
  return (
    <>
      {items.map((item, i) => (
        <ErrorCauseBoundary
          key={i}
          onError={(error) =>
            new Error(
              `A theme navbar item failed to render.
Please double-check the following navbar item (themeConfig.navbar.items) of your Docusaurus config:
${JSON.stringify(item, null, 2)}`,
              {cause: error},
            )
          }>
          <NavbarItem {...item} />
        </ErrorCauseBoundary>
      ))}
    </>
  );
}
function NavbarContentLayout({left, right}) {
  return (
    <div className="navbar__inner">
      <div className="navbar__items">{left}</div>
      <div className="navbar__items navbar__items--right">{right}</div>
    </div>
  );
}
export default function NavbarContent() {
  const mobileSidebar = useNavbarMobileSidebar();
  const items = useNavbarItems();
  const [leftItems, rightItems] = splitNavbarItems(items);
  const searchBarItem = items.find((item) => item.type === 'search');
  return (
    <NavbarContentLayout
      left={
        // TODO stop hardcoding items?
        <>
          {!mobileSidebar.disabled && <NavbarMobileSidebarToggle />}
          <NavbarLogo />
          <NavbarItems items={leftItems} />
        </>
      }
      right={
        // // TODO stop hardcoding items?
        // // Ask the user to add the respective navbar items => more flexible
        <>
          {/* <div style={{marginRight: "50px"}}>
            <Link href='https://buf.build/rigdev/rig'>
              API
            </Link>
          </div> */}
          <div style={{marginRight: "5px"}}>
            <Button width='35px' height='35px' textAlign='center' onClick={() => window.open("https://github.com/rigdev/rig")}> 
              <div style={{color: "var(--ifm-color-emphasis-700)", display: "block", justifyContent: "center", alignItems: "center"}}>
                  <DynamicBiIcon size={15} name={"BiLogoGithub"}/>
              </div>
            </Button>
          </div>
          <div style={{marginRight: "5px"}}>
            <Button width='35px' height='35px' textAlign='center' onClick={() => window.open("https://discord.gg/Tn5wmXMM2U")}> 
              <div style={{color: "var(--ifm-color-emphasis-700)", display: "block", justifyContent: "center", alignItems: "center"}}>
                  <DynamicBiIcon size={15} name={"BiLogoDiscordAlt"}/>
              </div>
            </Button>
          </div>
          <NavbarItems items={rightItems} />
          <NavbarColorModeToggle className={styles.colorModeToggle} />
          {!searchBarItem && (
            <NavbarSearch>
              <SearchBar />
            </NavbarSearch>
          )}
        </>
      }
    />
  );
}
