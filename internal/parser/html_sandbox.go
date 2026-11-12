package parser

import (
	"fmt"
	"strings"
)

type HTMLSandbox struct{}

func NewHTMLSandbox() *HTMLSandbox {
	return &HTMLSandbox{}
}

// WrapSandboxed embeds a client-side click-inspector bridge script into the HTML preview.
// When an element is clicked, it posts a message to the parent frame with element details and coordinates.
func (s *HTMLSandbox) WrapSandboxed(rawHTML string) string {
	bridgeScript := `
<script>
(function() {
  let inspectMode = true;
  let hoveredEl = null;

  const overlay = document.createElement('div');
  overlay.id = '__planmark_hover_overlay';
  overlay.style.position = 'fixed';
  overlay.style.pointerEvents = 'none';
  overlay.style.border = '2px dashed #3b82f6';
  overlay.style.backgroundColor = 'rgba(59, 130, 246, 0.15)';
  overlay.style.zIndex = '999999';
  overlay.style.display = 'none';
  overlay.style.transition = 'all 0.05s ease';
  document.documentElement.appendChild(overlay);

  window.addEventListener('message', (e) => {
    if (e.data && e.data.type === 'PLANMARK_SET_INSPECT') {
      inspectMode = !!e.data.enabled;
      if (!inspectMode) overlay.style.display = 'none';
    }
  });

  document.addEventListener('mousemove', (e) => {
    if (!inspectMode) return;
    const target = e.target;
    if (!target || target === overlay || target === document.documentElement || target === document.body) {
      overlay.style.display = 'none';
      return;
    }
    hoveredEl = target;
    const rect = target.getBoundingClientRect();
    overlay.style.display = 'block';
    overlay.style.top = rect.top + 'px';
    overlay.style.left = rect.left + 'px';
    overlay.style.width = rect.width + 'px';
    overlay.style.height = rect.height + 'px';
  }, true);

  document.addEventListener('click', (e) => {
    if (!inspectMode) return;
    e.preventDefault();
    e.stopPropagation();

    const target = e.target;
    if (!target || target === overlay) return;

    const rect = target.getBoundingClientRect();
    const selector = getCssSelector(target);

    window.parent.postMessage({
      type: 'PLANMARK_ELEMENT_PINNED',
      selector: selector,
      tagName: target.tagName.toLowerCase(),
      id: target.id || '',
      className: target.className || '',
      rect: {
        top: rect.top,
        left: rect.left,
        width: rect.width,
        height: rect.height,
        x: e.clientX,
        y: e.clientY
      }
    }, '*');
  }, true);

  function getCssSelector(el) {
    if (el.id) return '#' + el.id;
    let path = [];
    while (el && el.nodeType === Node.ELEMENT_NODE) {
      let selector = el.nodeName.toLowerCase();
      if (el.id) {
        selector += '#' + el.id;
        path.unshift(selector);
        break;
      } else {
        let sib = el, nth = 1;
        while (sib = sib.previousElementSibling) {
          if (sib.nodeName.toLowerCase() === selector) nth++;
        }
        if (nth !== 1) selector += ":nth-of-type(" + nth + ")";
      }
      path.unshift(selector);
      el = el.parentNode;
      if (path.length >= 3) break;
    }
    return path.join(' > ');
  }
})();
</script>
`

	if strings.Contains(rawHTML, "</body>") {
		return strings.Replace(rawHTML, "</body>", bridgeScript+"</body>", 1)
	}
	return rawHTML + bridgeScript
}

// SecurityHeaders returns standard protective headers for iframe serving
func (s *HTMLSandbox) SecurityHeaders() map[string]string {
	return map[string]string{
		"Content-Security-Policy": "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: https:; frame-ancestors 'self'",
		"X-Frame-Options":        "SAMEORIGIN",
	}
}

// GenerateWrapper generates an HTML doc if the payload is only a fragment
func (s *HTMLSandbox) EnsureDocument(snippet string) string {
	if strings.Contains(strings.ToLower(snippet), "<html") {
		return s.WrapSandboxed(snippet)
	}
	wrapped := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Component Sandbox</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="p-6 bg-slate-900 text-slate-100">
  %s
</body>
</html>`, snippet)
	return s.WrapSandboxed(wrapped)
}
