import 'package:flutter/material.dart';
import '../../design_system/components.dart';
import '../../design_system/icons.dart';
import '../../design_system/theme.dart';
import '../web/host_webview_page.dart';
import 'host_form_page.dart';
import 'host_store.dart';

enum _HostAction { edit, delete }

class HostsPage extends StatefulWidget {
  const HostsPage({super.key, required this.store});
  final HostStore store;

  @override
  State<HostsPage> createState() => _HostsPageState();
}

class _HostsPageState extends State<HostsPage> {
  List<SavedHost> _hosts = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final hosts = await widget.store.load();
      if (mounted) {
        setState(() {
          _hosts = hosts;
          _loading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Could not load saved web clients.';
          _loading = false;
        });
      }
    }
  }

  Future<void> _add() async {
    final changed = await Navigator.of(context).push<bool>(
      MaterialPageRoute(builder: (_) => HostFormPage(store: widget.store)),
    );
    if (mounted && changed == true) await _load();
  }

  Future<void> _edit(int index) async {
    final changed = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => HostFormPage(
          store: widget.store,
          host: _hosts[index],
          hostIndex: index,
        ),
      ),
    );
    if (mounted && changed == true) await _load();
  }

  Future<void> _delete(int index) async {
    final host = _hosts[index];
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Delete ${host.name}?'),
        content: const Text(
          'This removes the saved web client from this device.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            style: TextButton.styleFrom(
              foregroundColor: KlmColors.of(context).danger,
            ),
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (!mounted || confirmed != true) return;
    try {
      await widget.store.removeAt(index);
      if (mounted) await _load();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Could not delete this web client.')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = KlmColors.of(context);
    return KlmPage(
      title: const KlmBrand(),
      actions: [
        KlmIconButton(
          label: 'Add web client',
          onPressed: _loading || _error != null ? null : _add,
          child: const KlmIcon(KlmSymbol.plus),
        ),
      ],
      child: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    KlmError(_error!),
                    const SizedBox(height: 24),
                    KlmButton(
                      label: 'Retry',
                      icon: KlmSymbol.retry,
                      onPressed: _load,
                    ),
                  ],
                ),
              ),
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Padding(
                  padding: EdgeInsets.fromLTRB(24, 28, 24, 16),
                  child: Text(
                    'Web clients',
                    style: TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.w600,
                      letterSpacing: -.5,
                    ),
                  ),
                ),
                Expanded(
                  child: _hosts.isEmpty
                      ? Center(
                          child: Padding(
                            padding: const EdgeInsets.all(32),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(18),
                                  decoration: BoxDecoration(
                                    color: colors.subtle,
                                    borderRadius: BorderRadius.circular(16),
                                  ),
                                  child: KlmIcon(
                                    KlmSymbol.server,
                                    size: 28,
                                    color: colors.muted,
                                  ),
                                ),
                                const SizedBox(height: 24),
                                const Text(
                                  'No web clients yet',
                                  style: TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                                const SizedBox(height: 24),
                                KlmButton(
                                  label: 'Add web client',
                                  icon: KlmSymbol.plus,
                                  onPressed: _add,
                                ),
                              ],
                            ),
                          ),
                        )
                      : ListView.separated(
                          padding: const EdgeInsets.fromLTRB(24, 0, 24, 24),
                          itemCount: _hosts.length,
                          separatorBuilder: (_, _) =>
                              const SizedBox(height: 12),
                          itemBuilder: (context, index) {
                            final host = _hosts[index];
                            return KlmCard(
                              onTap: () => Navigator.of(context).push(
                                MaterialPageRoute(
                                  builder: (_) => HostWebViewPage(host: host),
                                ),
                              ),
                              child: Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.all(12),
                                    decoration: BoxDecoration(
                                      color: colors.subtle,
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: host.emoji == null
                                        ? KlmIcon(
                                            KlmSymbol.server,
                                            color: colors.muted,
                                          )
                                        : Text(
                                            host.emoji!,
                                            style: const TextStyle(
                                              fontSize: 22,
                                            ),
                                          ),
                                  ),
                                  const SizedBox(width: 16),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          host.name,
                                          maxLines: 2,
                                          overflow: TextOverflow.ellipsis,
                                          style: TextStyle(
                                            fontSize: 15,
                                            fontWeight: FontWeight.w500,
                                            color: colors.heading,
                                          ),
                                        ),
                                        const SizedBox(height: 4),
                                        Text(
                                          host.url.toString(),
                                          maxLines: 2,
                                          overflow: TextOverflow.ellipsis,
                                          style: TextStyle(
                                            fontSize: 13,
                                            color: colors.muted,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  PopupMenuButton<_HostAction>(
                                    tooltip: 'Web client actions',
                                    icon: KlmIcon(
                                      KlmSymbol.moreVertical,
                                      color: colors.faint,
                                    ),
                                    onSelected: (action) {
                                      switch (action) {
                                        case _HostAction.edit:
                                          _edit(index);
                                          break;
                                        case _HostAction.delete:
                                          _delete(index);
                                          break;
                                      }
                                    },
                                    itemBuilder: (context) => [
                                      const PopupMenuItem(
                                        value: _HostAction.edit,
                                        child: Row(
                                          children: [
                                            KlmIcon(KlmSymbol.edit, size: 18),
                                            SizedBox(width: 12),
                                            Text('Edit'),
                                          ],
                                        ),
                                      ),
                                      PopupMenuItem(
                                        value: _HostAction.delete,
                                        child: Row(
                                          children: [
                                            KlmIcon(
                                              KlmSymbol.trash,
                                              size: 18,
                                              color: colors.danger,
                                            ),
                                            const SizedBox(width: 12),
                                            Text(
                                              'Delete',
                                              style: TextStyle(
                                                color: colors.danger,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            );
                          },
                        ),
                ),
              ],
            ),
    );
  }
}
