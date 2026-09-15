import 'dart:io';

/// Builds the web client URL. Only literal IPs get the KLM frontend default.
/// Bare domains use HTTPS; explicit HTTP(S) schemes and ports are preserved.
Uri parseHostAddress(String input) {
  var value = input.trim();
  String? scheme;
  final protocol = RegExp(
    r'^(https?)://',
    caseSensitive: false,
  ).firstMatch(value);
  if (protocol != null) {
    scheme = protocol.group(1)!.toLowerCase();
    value = value.substring(protocol.end);
  }
  if (value.endsWith('/')) value = value.substring(0, value.length - 1);
  if (value.isEmpty ||
      RegExp(r'[\s/@?#]').hasMatch(value) ||
      value.contains('://')) {
    throw const FormatException('Enter a web client URL without a path.');
  }

  String host = value;
  int? port;
  if (value.startsWith('[')) {
    final match = RegExp(r'^\[([^\]]+)\](?::([0-9]+))?$').firstMatch(value);
    if (match == null ||
        InternetAddress.tryParse(match.group(1)!)?.type !=
            InternetAddressType.IPv6) {
      throw const FormatException(
        'Enter a valid IPv6 address, using [IP]:port for a custom port.',
      );
    }
    host = match.group(1)!;
    if (match.group(2) != null) port = int.tryParse(match.group(2)!);
    if (match.group(2) != null && port == null) {
      throw const FormatException('Port must be between 1 and 65535.');
    }
  } else if (InternetAddress.tryParse(value) == null && value.contains(':')) {
    final match = RegExp(r'^([^:]+):([0-9]+)$').firstMatch(value);
    if (match == null) {
      throw const FormatException('Enter a valid web client URL and port.');
    }
    host = match.group(1)!;
    port = int.tryParse(match.group(2)!);
    if (port == null) {
      throw const FormatException('Port must be between 1 and 65535.');
    }
  }
  if (port != null && (port < 1 || port > 65535)) {
    throw const FormatException('Port must be between 1 and 65535.');
  }
  final ip = InternetAddress.tryParse(host);
  if (ip == null) {
    final domain = host.endsWith('.')
        ? host.substring(0, host.length - 1)
        : host;
    final label = RegExp(r'^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$');
    if (domain.length > 253 ||
        RegExp(r'^[0-9.]+$').hasMatch(domain) ||
        domain.split('.').any((part) => !label.hasMatch(part))) {
      throw const FormatException('Enter a valid web client URL.');
    }
  }
  return Uri(
    scheme: scheme ?? (ip == null ? 'https' : 'http'),
    host: host,
    port: port ?? (ip == null ? null : 7332),
  );
}
