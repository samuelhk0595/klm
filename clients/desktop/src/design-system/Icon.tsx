import type { LucideIcon } from 'lucide-react';

export type IconSize = 10 | 12 | 14 | 16 | 18 | 20 | 24;

export type IconProps = {
  glyph: LucideIcon;
  size?: IconSize;
  className?: string;
};

export function Icon({ glyph: Glyph, size = 16, className }: IconProps) {
  return <Glyph aria-hidden="true" className={className} size={size} strokeWidth={1.8} style={{ width: size, height: size }} />;
}
