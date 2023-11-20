import React from "react";

import * as Icons from "react-icons/bi";
import * as SimpleIcons from "react-icons/si";
import * as TablerIcons from "react-icons/tb";

/* Your icon name from database data can now be passed as prop */
const DynamicBiIcon = ({ name, size }) => {
  let IconComponent = Icons[name];
  if (!IconComponent) {
    IconComponent = SimpleIcons[name];
  }
  if (!IconComponent) {
    IconComponent = TablerIcons[name];
  }

  if (!IconComponent) {
    // Return a default one
    return <Icons.BiAbacus size={size} />;
  }

  return <IconComponent size={size} />;
};

export default DynamicBiIcon;
