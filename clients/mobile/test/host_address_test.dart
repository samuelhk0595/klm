import 'package:flutter_test/flutter_test.dart';
import 'package:klm_harness/features/hosts/host_address.dart';

void main() {
  test('only literal IPs receive port 7332 when no port was supplied', () {
    expect(
      parseHostAddress('192.168.1.10').toString(),
      'http://192.168.1.10:7332',
    );
    expect(parseHostAddress('http://10.0.2.2').port, 7332);
    expect(
      parseHostAddress('https://192.168.1.10').toString(),
      'https://192.168.1.10:7332',
    );
    expect(parseHostAddress('::1').toString(), 'http://[::1]:7332');
    expect(parseHostAddress('[2001:db8::1]').port, 7332);
    expect(
      parseHostAddress('klm.example.com').toString(),
      'https://klm.example.com',
    );
    expect(parseHostAddress('https://klm.example.com/').hasPort, false);
    expect(
      parseHostAddress('http://klm.example.com').toString(),
      'http://klm.example.com',
    );
  });

  test('explicit ports and protocols win, including standard ports', () {
    expect(parseHostAddress('192.168.1.10:8080').port, 8080);
    expect(parseHostAddress('http://192.168.1.10:80').port, 80);
    expect(parseHostAddress('https://192.168.1.10:443').port, 443);
    expect(parseHostAddress('[::1]:8443').port, 8443);
    expect(parseHostAddress('klm.example.com:8443').port, 8443);
    expect(
      parseHostAddress('http://klm.example.com:7332').toString(),
      'http://klm.example.com:7332',
    );
  });

  test('invalid addresses are not saved as plausible hosts', () {
    for (final value in [
      '',
      '999.0.0.1',
      '192.168.1.10:',
      '192.168.1.10:0',
      'klm.example.com:65536',
      'ftp://klm.example.com',
      'https://user:pass@klm.example.com',
      'https://klm.example.com/path',
      'https://klm.example.com?token=secret',
      '[::1',
      'bad host',
    ]) {
      expect(
        () => parseHostAddress(value),
        throwsFormatException,
        reason: value,
      );
    }
  });
}
