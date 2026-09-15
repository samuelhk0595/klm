import 'package:flutter/material.dart';
import 'design_system/theme.dart';
import 'features/hosts/hosts_page.dart';
import 'features/hosts/host_store.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(KlmHarnessApp(store: HostStore()));
}

class KlmHarnessApp extends StatelessWidget {
  const KlmHarnessApp({super.key, required this.store});

  final HostStore store;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'KLM Harness',
      debugShowCheckedModeBanner: false,
      theme: KlmTheme.light,
      darkTheme: KlmTheme.dark,
      home: HostsPage(store: store),
    );
  }
}
