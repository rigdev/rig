import React from "react"

type ButtonProps = {
  className?: string
  onClick?: React.MouseEventHandler<HTMLButtonElement>
  disabled?: boolean
  height?: string
  width?: string
  backgroundColor?: string
  border?: string
  borderRadius?: string
  textAlign?: "center" | "left" | "right"
  padding?: string
} & React.HTMLAttributes<HTMLButtonElement>

const Button: React.FC<ButtonProps> = ({
  className = "",
  height = "35px",
  width = "110px",
  backgroundColor = "var(--ifm-color-emphasis-100)",
  border = "1px solid var(--ifm-color-emphasis-200)",
  borderRadius = "8px",
  textAlign = "center",
  padding = "0px",
  onClick,
  children,
  ...props
}) => {
  return (
    <button style={{
        borderRadius: borderRadius,
        border: border,
        whiteSpace: "nowrap",
        backgroundColor: backgroundColor,
        textAlign: textAlign,
        padding: padding,
        height: height,
        width: width,
        fontWeight: 500,
    }}
      onClick={onClick}
      onMouseOver={(e) => {
        e.currentTarget.style.opacity = "0.6"
        e.currentTarget.style.cursor = "pointer"
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.opacity = "1"
      }}
      {...props}
    >
      {children}
    </button>
  )
}

export default Button