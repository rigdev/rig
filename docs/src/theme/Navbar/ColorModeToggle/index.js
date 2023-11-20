import React from 'react';
import {useColorMode, useThemeConfig} from '@docusaurus/theme-common';
import Button from '../../../components/Button';
import DynamicBiIcon from '../../../components/DynamicBiIcon';
export default function NavbarColorModeToggle({className}) {
  const navbarStyle = useThemeConfig().navbar.style;
  const disabled = useThemeConfig().colorMode.disableSwitch;
  const {colorMode, setColorMode} = useColorMode();
  if (disabled) {
    return null;
  }

  return (
    <Button onClick={() => setColorMode(colorMode === "light" ? "dark" : "light")} width='35px' height='35px' textAlign='center'> 
      <div style={{color: "var(--ifm-color-emphasis-700)", display: "block", justifyContent: "center", alignItems: "center"}}>
        {colorMode === "dark" && (
          <DynamicBiIcon size={15} name={"BiMoon"}/>
        )}
        {colorMode === "light" && (
          <DynamicBiIcon size={15} name={"BiSun"}/>
        )}
      </div>
    </Button>
  );
}