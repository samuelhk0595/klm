import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

enum KlmSymbol {
  plus,
  arrowLeft,
  chevronRight,
  server,
  retry,
  alert,
  moreVertical,
  edit,
  trash,
}

/// The same Lucide outlines used by the web client; 24-unit view box, 2px stroke.
class KlmIcon extends StatelessWidget {
  const KlmIcon(this.symbol, {super.key, this.size = 20, this.color});
  final KlmSymbol symbol;
  final double size;
  final Color? color;

  static const _paths = {
    KlmSymbol.plus: '<path d="M12 5v14M5 12h14"/>',
    KlmSymbol.arrowLeft: '<path d="m12 19-7-7 7-7M5 12h14"/>',
    KlmSymbol.chevronRight: '<path d="m9 18 6-6-6-6"/>',
    KlmSymbol.server:
        '<rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/><path d="M6 6.5h.01M6 17.5h.01"/>',
    KlmSymbol.retry:
        '<path d="M3 11a9 9 0 0 1 15-6.7L21 7M21 3v4h-4M21 13a9 9 0 0 1-15 6.7L3 17M7 17H3v4"/>',
    KlmSymbol.alert:
        '<circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/>',
    KlmSymbol.moreVertical:
        '<circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/>',
    KlmSymbol.edit:
        '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/>',
    KlmSymbol.trash:
        '<path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6M10 11v5M14 11v5"/>',
  };

  @override
  Widget build(BuildContext context) => SvgPicture.string(
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="black" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${_paths[symbol]}</svg>',
    width: size,
    height: size,
    colorFilter: ColorFilter.mode(
      color ?? IconTheme.of(context).color ?? Colors.black,
      BlendMode.srcIn,
    ),
    excludeFromSemantics: true,
  );
}
