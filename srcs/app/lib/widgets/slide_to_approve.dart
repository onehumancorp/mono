import 'package:flutter/material.dart';

/// A custom gesture-based slider for high-stakes approvals.
class SlideToApprove extends StatefulWidget {
  final VoidCallback onApprove;
  final VoidCallback onReject;
  final bool disabled;

  const SlideToApprove({
    super.key,
    required this.onApprove,
    required this.onReject,
    this.disabled = false,
  });

  @override
  State<SlideToApprove> createState() => _SlideToApproveState();
}

class _SlideToApproveState extends State<SlideToApprove> {
  double _position = 0.0;
  bool _isDragging = false;
  static const double _maxDrag = 240.0;
  static const double _thumbWidth = 48.0;

  void _handleDragUpdate(DragUpdateDetails details) {
    if (widget.disabled) return;
    setState(() {
      _isDragging = true;
      _position += details.delta.dx;
      if (_position < 0) _position = 0;
      if (_position > _maxDrag) _position = _maxDrag;
    });
  }

  void _handleDragEnd(DragEndDetails details) {
    if (widget.disabled) return;
    setState(() {
      _isDragging = false;
      if (_position > _maxDrag * 0.9) {
        _position = _maxDrag;
        widget.onApprove();
      } else {
        _position = 0.0;
      }
    });
  }

  @override
  void didUpdateWidget(SlideToApprove oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Reset if was disabled and now enabled (e.g. error) or if was enabled and now disabled (start processing)
    if (!widget.disabled && _position == _maxDrag) {
      setState(() {
        _position = 0.0;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          height: 56,
          width: _maxDrag + _thumbWidth,
          decoration: BoxDecoration(
            color: colors.surfaceContainerHighest.withOpacity(0.3),
            borderRadius: BorderRadius.circular(28),
            border: Border.all(
              color: colors.outlineVariant.withOpacity(0.2),
            ),
          ),
          child: Stack(
            children: [
              // Track Text
              Center(
                child: widget.disabled && _position == _maxDrag
                    ? Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                            ),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            'Authorizing...',
                            style: TextStyle(
                              color: colors.primary.withOpacity(0.7),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      )
                    : Text(
                        'Slide to Approve',
                        style: TextStyle(
                          color: colors.onSurfaceVariant.withOpacity(0.5),
                          fontWeight: FontWeight.w500,
                          letterSpacing: 0.5,
                        ),
                      ),
              ),

              // Progress Fill
              Positioned(
                left: 0,
                top: 4,
                bottom: 4,
                child: Container(
                  width: _position + (_thumbWidth / 2),
                  decoration: BoxDecoration(
                    color: colors.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(24),
                  ),
                ),
              ),

              // Thumb
              AnimatedPositioned(
                duration: _isDragging ? Duration.zero : const Duration(milliseconds: 200),
                curve: Curves.easeOut,
                left: _position + 4,
                top: 4,
                bottom: 4,
                child: GestureDetector(
                  onHorizontalDragUpdate: _handleDragUpdate,
                  onHorizontalDragEnd: _handleDragEnd,
                  child: Container(
                    width: _thumbWidth,
                    decoration: BoxDecoration(
                      color: widget.disabled && _position == _maxDrag
                          ? Colors.green.withOpacity(0.2)
                          : colors.primary,
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: [
                        BoxShadow(
                          color: colors.shadow.withOpacity(0.1),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Center(
                      child: widget.disabled && _position == _maxDrag
                          ? const Icon(Icons.check, color: Colors.green, size: 24)
                          : const Icon(Icons.chevron_right, color: Colors.white, size: 24),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: TextButton(
            onPressed: widget.disabled ? null : widget.onReject,
            style: TextButton.styleFrom(
              backgroundColor: Colors.red.withOpacity(0.05),
              foregroundColor: Colors.red,
              side: BorderSide(color: Colors.red.withOpacity(0.1)),
              padding: const EdgeInsets.symmetric(vertical: 12),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: const Text('Reject Request'),
          ),
        ),
      ],
    );
  }
}
