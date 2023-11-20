import React from "react"
import DynamicBiIcon from "../DynamicBiIcon"

type IconBoxProps = {
  className?: string
  logo?: string
  size?: number
  gap?: number
} & React.HTMLAttributes<HTMLButtonElement>

const IconBox: React.FC<IconBoxProps> = ({
  className = "",
  logo,
  size = 45,
  gap = 10,
  onClick,
  ...props
}) => {
  return (
      <div
      style={{
        background: "var(--ifm-color-emphasis-100)",
        border: "1px solid var(--ifm-color-emphasis-200)",
        width: `${size}px`,
        height: `${size}px`,
        borderRadius: "8px",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        }}>
        <div style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: `${size-gap}px`,
          width: `${size-gap}px`,
          backgroundColor: "var(--ifm-color-emphasis-200)",
          borderRadius: "6px",
          color: "var(--ifm-color-emphasis-800)"
        }}>
          <DynamicBiIcon size={20} name={logo} />
        </div>
      </div>
  )
}

export default IconBox