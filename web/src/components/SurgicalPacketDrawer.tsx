import React, { useState } from 'react';
import { Session, Annotation } from '../types';
import { X, Copy, Check, Download, Zap, GitPullRequest } from 'lucide-react';

interface SurgicalPacketDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  session: Session | null;
  annotations: Annotation[];
}

export const SurgicalPacketDrawer: React.FC<SurgicalPacketDrawerProps> = ({
  isOpen,
  onClose,
  session,
  annotations,
}) => {
  const [copied, setCopied] = useState(false);
  const [copiedPR, setCopiedPR] = useState(false);

  if (!isOpen || !session) return null;

  // Compile surgical patch markdown
  const compilePacket = () => {
    let out = `<!-- planmark:feedback-packet v1 -->\n`;
    out += `# Human Review Feedback: ${session.title}\n`;
    out += `Session: ${session.id}\n`;
    out += `Agent: ${session.agent}\n`;
    out += `Timestamp: ${new Date().toISOString()}\n`;
    out += `Status: ${annotations.length > 0 ? 'changes_requested' : 'approved'}\n\n`;

    if (annotations.length === 0) {
      out += `> Approved as specified. No modifications requested.\n`;
    }

    annotations.forEach((ann, i) => {
      const idx = i + 1;
      switch (ann.type) {
        case 'replacement':
          out += `## Annotation ${idx}: Text Replacement (Lines ${ann.startLine}-${ann.endLine})\n`;
          out += `<<<<<<< ORIGINAL\n${(ann.original || '').trim()}\n=======\n${(ann.proposed || '').trim()}\n>>>>>>> RATIONALE\n${(ann.rationale || '').trim()}\n\n`;
          break;
        case 'suggestion':
          out += `## Annotation ${idx}: Code Suggestion (${ann.target || session.artifact.filePath || 'code'}:${ann.startLine || 1})\n`;
          out += `\`\`\`suggestion\n${(ann.proposed || '').trim()}\n\`\`\`\n`;
          if (ann.comment) {
            out += `*Comment: ${ann.comment.trim()}*\n`;
          }
          out += `\n`;
          break;
        case 'deletion':
          out += `## Annotation ${idx}: Deletion (Lines ${ann.startLine}-${ann.endLine})\n`;
          out += `<<<<<<< ORIGINAL\n${(ann.original || '').trim()}\n======= DELETION\n>>>>>>> RATIONALE\n${(ann.rationale || '').trim()}\n\n`;
          break;
        case 'comment':
          out += `## Annotation ${idx}: Comment (Line ${ann.startLine || 1})\n`;
          out += `> ${(ann.comment || '').trim()}\n\n`;
          break;
        case 'pin':
          out += `## Annotation ${idx}: Visual Pin (${ann.target || 'element'})\n`;
          if (ann.coordinates) {
            out += `- Target: \`${ann.coordinates.selector}\`\n`;
            out += `- Coordinates: (x: ${ann.coordinates.x.toFixed(1)}, y: ${ann.coordinates.y.toFixed(1)}, viewport: ${ann.coordinates.viewport}px)\n`;
          }
          out += `- Note: "${(ann.comment || '').trim()}"\n\n`;
          break;
      }
    });

    out += `<!-- planmark:end-packet -->\n`;
    return out;
  };

  // Compile formatted Markdown specifically for GitHub PR review comments
  const compileGitHubPRMarkdown = () => {
    let out = `### 🔍 Review Feedback on \`${session.artifact.filePath || session.title}\`\n\n`;
    annotations.forEach((ann, idx) => {
      out += `#### ${idx + 1}. ${ann.type === 'suggestion' ? 'Code Suggestion' : 'Revision'} (${ann.target || 'Line ' + (ann.startLine || 1)})\n`;
      if (ann.proposed) {
        out += `\`\`\`suggestion\n${ann.proposed.trim()}\n\`\`\`\n`;
      }
      if (ann.rationale || ann.comment) {
        out += `> ${ann.rationale || ann.comment}\n\n`;
      }
    });
    return out;
  };

  const rawPacket = compilePacket();

  // Token estimator
  let origChars = 0;
  let propChars = 0;
  annotations.forEach((a) => {
    origChars += (a.original || '').length;
    propChars += (a.proposed || '').length;
  });
  const origTokens = Math.round(origChars / 4);
  const propTokens = Math.round(propChars / 4);
  const tokenDelta = propTokens - origTokens;

  const handleCopy = () => {
    navigator.clipboard.writeText(rawPacket);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCopyPR = () => {
    navigator.clipboard.writeText(compileGitHubPRMarkdown());
    setCopiedPR(true);
    setTimeout(() => setCopiedPR(false), 2000);
  };

  const handleDownload = () => {
    const blob = new Blob([rawPacket], { type: 'text/markdown;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `planmark_feedback_${session.id}.md`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-2xl bg-[#0d1424] border-l border-gray-800 flex flex-col h-full shadow-2xl animate-in slide-in-from-right duration-200">
        {/* Header */}
        <div className="h-14 border-b border-gray-800 px-5 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="w-4 h-4 text-blue-400" />
            <h3 className="font-semibold text-slate-100 text-sm">Deterministic Surgical Feedback Packet</h3>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 text-gray-400 hover:text-white rounded hover:bg-gray-800 transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Token Impact Estimator Bar */}
        <div className="bg-[#111a2e] px-5 py-2.5 border-b border-gray-800 flex items-center justify-between text-xs">
          <div className="flex items-center gap-3">
            <span className="text-gray-400">Context Impact:</span>
            <span className="font-mono text-gray-300">
              Orig: <strong className="text-slate-200">{origTokens}</strong> tok
            </span>
            <span className="font-mono text-gray-300">
              Prop: <strong className="text-slate-200">{propTokens}</strong> tok
            </span>
            <span
              className={`font-mono font-medium px-1.5 py-0.5 rounded ${
                tokenDelta > 0
                  ? 'bg-amber-950/60 text-amber-400 border border-amber-800/40'
                  : 'bg-emerald-950/60 text-emerald-400 border border-emerald-800/40'
              }`}
            >
              Delta: {tokenDelta > 0 ? `+${tokenDelta}` : tokenDelta} tok
            </span>
          </div>

          <div className="flex items-center gap-1.5">
            <button
              onClick={handleCopyPR}
              className="flex items-center gap-1 px-2.5 py-1 rounded bg-gray-800 hover:bg-gray-700 text-slate-200 transition"
              title="Copy formatted for GitHub Pull Request review comments"
            >
              {copiedPR ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <GitPullRequest className="w-3.5 h-3.5 text-emerald-400" />}
              <span>{copiedPR ? 'PR Copied' : 'GitHub PR'}</span>
            </button>

            <button
              onClick={handleCopy}
              className="flex items-center gap-1 px-2.5 py-1 rounded bg-gray-800 hover:bg-gray-700 text-slate-200 transition"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copied ? 'Copied' : 'Copy'}</span>
            </button>

            <button
              onClick={handleDownload}
              className="flex items-center gap-1 px-2.5 py-1 rounded bg-gray-800 hover:bg-gray-700 text-slate-200 transition"
            >
              <Download className="w-3.5 h-3.5" />
              <span>Export</span>
            </button>
          </div>
        </div>

        {/* Content Preview */}
        <div className="flex-1 p-5 overflow-y-auto font-mono text-xs text-slate-200 bg-[#090e1a]">
          <pre className="whitespace-pre-wrap leading-relaxed select-all selection:bg-blue-600">
            {rawPacket}
          </pre>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-gray-800 bg-[#0c1220] flex items-center justify-between text-xs text-gray-400">
          <span>This packet will be piped directly back into the calling agent process.</span>
          <button
            onClick={onClose}
            className="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded font-medium transition"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
};
