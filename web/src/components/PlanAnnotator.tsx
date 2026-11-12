import React, { useState, useRef, useEffect } from 'react';
import { Annotation } from '../types';
import { MessageSquare, Edit3, Trash2, Check, X, CornerDownRight, CheckSquare, Square, SlidersHorizontal } from 'lucide-react';
import { VoiceDictationButton } from './VoiceDictationButton';
import { GuardrailChips } from './GuardrailChips';
import { computeWordDiff } from '../utils/wordDiff';

interface PlanAnnotatorProps {
  content: string;
  annotations: Annotation[];
  onAddAnnotation: (ann: Omit<Annotation, 'id' | 'createdAt'>) => void;
  onDeleteAnnotation: (id: string) => void;
}

interface SelectionState {
  startLine: number;
  endLine: number;
  selectedText: string;
  x: number;
  y: number;
}

export const PlanAnnotator: React.FC<PlanAnnotatorProps> = ({
  content,
  annotations,
  onAddAnnotation,
  onDeleteAnnotation,
}) => {
  const [lines, setLines] = useState<string[]>(content.split('\n'));
  const [selection, setSelection] = useState<SelectionState | null>(null);
  const [activeDialog, setActiveDialog] = useState<'replacement' | 'comment' | null>(null);
  const [proposedText, setProposedText] = useState('');
  const [rationale, setRationale] = useState('');
  const [commentText, setCommentText] = useState('');
  const [density, setDensity] = useState<'comfortable' | 'compact'>('comfortable');
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setLines(content.split('\n'));
  }, [content]);

  // Clear selection when clicking outside
  useEffect(() => {
    const handleDocumentClick = (e: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        if (!activeDialog) setSelection(null);
      }
    };
    document.addEventListener('mousedown', handleDocumentClick);
    return () => document.removeEventListener('mousedown', handleDocumentClick);
  }, [activeDialog]);

  const handleMouseUp = () => {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed || sel.rangeCount === 0) {
      if (!activeDialog) setSelection(null);
      return;
    }

    const text = sel.toString().trim();
    if (!text) return;

    const range = sel.getRangeAt(0);
    const rect = range.getBoundingClientRect();

    let startLine = 1;
    let endLine = 1;

    const startNode = range.startContainer.parentElement?.closest('[data-line]');
    const endNode = range.endContainer.parentElement?.closest('[data-line]');

    if (startNode) {
      startLine = parseInt(startNode.getAttribute('data-line') || '1', 10);
    }
    if (endNode) {
      endLine = parseInt(endNode.getAttribute('data-line') || '1', 10);
    } else {
      endLine = startLine;
    }

    if (startLine > endLine) {
      const tmp = startLine;
      startLine = endLine;
      endLine = tmp;
    }

    setSelection({
      startLine,
      endLine,
      selectedText: text,
      x: rect.left + rect.width / 2,
      y: rect.top - 10,
    });
    setProposedText(text);
  };

  const submitReplacement = () => {
    if (!selection) return;
    onAddAnnotation({
      type: 'replacement',
      startLine: selection.startLine,
      endLine: selection.endLine,
      original: selection.selectedText,
      proposed: proposedText,
      rationale: rationale || 'Improve implementation precision',
    });
    resetForms();
  };

  const submitComment = () => {
    if (!selection) return;
    onAddAnnotation({
      type: 'comment',
      startLine: selection.startLine,
      endLine: selection.endLine,
      original: selection.selectedText,
      comment: commentText,
    });
    resetForms();
  };

  const submitDeletion = () => {
    if (!selection) return;
    onAddAnnotation({
      type: 'deletion',
      startLine: selection.startLine,
      endLine: selection.endLine,
      original: selection.selectedText,
      rationale: rationale || 'Remove redundant specification clause',
    });
    resetForms();
  };

  const resetForms = () => {
    setSelection(null);
    setActiveDialog(null);
    setProposedText('');
    setRationale('');
    setCommentText('');
    window.getSelection()?.removeAllRanges();
  };

  // Toggle markdown task checkbox line
  const handleToggleCheckbox = (lineNum: number, currentText: string, isChecked: boolean) => {
    const updatedLine = isChecked
      ? currentText.replace(/\[[xX]\]/, '[ ]')
      : currentText.replace(/\[\s*\]/, '[x]');

    setLines((prev) => {
      const next = [...prev];
      next[lineNum - 1] = updatedLine;
      return next;
    });

    onAddAnnotation({
      type: 'comment',
      startLine: lineNum,
      endLine: lineNum,
      comment: isChecked
        ? `Task marked as uncompleted / pending verification`
        : `Task approved and verified as completed`,
    });
  };

  const getLineAnnotation = (lineNum: number) => {
    return annotations.find(
      (a) => a.startLine && a.endLine && lineNum >= a.startLine && lineNum <= a.endLine
    );
  };

  return (
    <div className="flex h-[calc(100vh-3.5rem)] relative overflow-hidden" ref={containerRef}>
      {/* Main Document Viewer */}
      <div
        className="flex-1 overflow-y-auto p-4 md:p-6 font-mono text-sm leading-relaxed"
        onMouseUp={handleMouseUp}
      >
        <div className="max-w-4xl mx-auto bg-[#0d1424] border border-gray-800 rounded-lg p-5 shadow-xl">
          {/* Header Bar */}
          <div className="mb-4 pb-3 border-b border-gray-800 flex items-center justify-between text-xs text-gray-400">
            <span>Highlight text to leave comments, surgical replacement blocks, or strikes.</span>
            <div className="flex items-center gap-3">
              <span className="font-mono">{lines.length} lines</span>
              <button
                onClick={() => setDensity((d) => (d === 'comfortable' ? 'compact' : 'comfortable'))}
                className="flex items-center gap-1 px-2 py-0.5 rounded bg-gray-800 hover:bg-gray-700 text-slate-300 transition"
                title="Toggle visual line density"
              >
                <SlidersHorizontal className="w-3 h-3" />
                <span className="capitalize text-[11px]">{density}</span>
              </button>
            </div>
          </div>

          {/* Line Listing */}
          <div className="divide-y divide-gray-800/40">
            {lines.map((line, idx) => {
              const lineNum = idx + 1;
              const ann = getLineAnnotation(lineNum);
              let bgClass = 'hover:bg-gray-800/40';

              if (ann) {
                if (ann.type === 'replacement') bgClass = 'bg-amber-950/30 border-l-2 border-amber-500';
                else if (ann.type === 'deletion')
                  bgClass = 'bg-rose-950/30 border-l-2 border-rose-500 line-through text-gray-500';
                else if (ann.type === 'comment') bgClass = 'bg-blue-950/30 border-l-2 border-blue-500';
              }

              // Check for task checklist markdown: - [ ] or - [x]
              const taskMatch = line.match(/^(\s*[-*]\s+)\[([ xX])\]\s*(.*)$/);
              const isTask = !!taskMatch;
              const isChecked = taskMatch ? taskMatch[2].toLowerCase() === 'x' : false;

              return (
                <div
                  key={lineNum}
                  data-line={lineNum}
                  className={`flex items-start rounded transition ${
                    density === 'compact' ? 'py-0 px-1 text-xs' : 'py-0.5 px-2 text-sm'
                  } ${bgClass}`}
                >
                  <span className="w-10 shrink-0 text-right pr-4 text-xs text-gray-400 select-none font-mono">
                    {lineNum}
                  </span>

                  {isTask ? (
                    <div className="flex items-center gap-2 flex-1">
                      <button
                        type="button"
                        onClick={() => handleToggleCheckbox(lineNum, line, isChecked)}
                        className="text-blue-400 hover:text-blue-300 transition shrink-0"
                      >
                        {isChecked ? (
                          <CheckSquare className="w-4 h-4 text-emerald-400" />
                        ) : (
                          <Square className="w-4 h-4 text-gray-400" />
                        )}
                      </button>
                      <span
                        className={`whitespace-pre-wrap break-words ${
                          isChecked ? 'line-through text-gray-400' : 'text-slate-200'
                        }`}
                      >
                        {taskMatch[3]}
                      </span>
                    </div>
                  ) : (
                    <span className="flex-1 whitespace-pre-wrap break-words text-slate-200">
                      {line || '\u00A0'}
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Floating Selection Toolbar */}
      {selection && !activeDialog && (
        <div
          className="fixed z-50 transform -translate-x-1/2 -translate-y-full bg-[#1e293b] border border-gray-700 rounded-lg shadow-2xl p-1 flex items-center gap-1 text-xs"
          style={{ left: `${selection.x}px`, top: `${selection.y}px` }}
        >
          <button
            onClick={() => setActiveDialog('replacement')}
            className="flex items-center gap-1.5 px-2.5 py-1.5 hover:bg-blue-600 rounded text-slate-200 hover:text-white transition"
          >
            <Edit3 className="w-3.5 h-3.5 text-blue-400" />
            <span>Suggest Replacement</span>
          </button>
          <button
            onClick={() => setActiveDialog('comment')}
            className="flex items-center gap-1.5 px-2.5 py-1.5 hover:bg-blue-600 rounded text-slate-200 hover:text-white transition"
          >
            <MessageSquare className="w-3.5 h-3.5 text-blue-400" />
            <span>Comment</span>
          </button>
          <button
            onClick={submitDeletion}
            className="flex items-center gap-1.5 px-2.5 py-1.5 hover:bg-rose-600 rounded text-slate-200 hover:text-white transition"
          >
            <Trash2 className="w-3.5 h-3.5 text-rose-400" />
            <span>Strike</span>
          </button>
        </div>
      )}

      {/* Replacement Modal */}
      {selection && activeDialog === 'replacement' && (
        <div
          className="fixed z-50 transform -translate-x-1/2 bg-[#172033] border border-blue-500/50 rounded-xl shadow-2xl p-4 w-[420px] text-xs text-slate-200 animate-in fade-in zoom-in-95 duration-150"
          style={{ left: `${selection.x}px`, top: `${Math.max(20, selection.y - 140)}px` }}
        >
          <div className="flex items-center justify-between pb-2 mb-2 border-b border-gray-700">
            <span className="font-semibold text-blue-400 flex items-center gap-1.5">
              <Edit3 className="w-3.5 h-3.5" />
              Suggest Surgical Replacement (L{selection.startLine}-{selection.endLine})
            </span>
            <button onClick={resetForms} className="text-gray-400 hover:text-white">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="space-y-3">
            <div>
              <label className="block text-[11px] text-gray-400 mb-1">Proposed Text:</label>
              <textarea
                value={proposedText}
                onChange={(e) => setProposedText(e.target.value)}
                rows={3}
                className="w-full bg-[#0b101c] border border-gray-700 rounded p-2 text-xs font-mono text-slate-100 focus:outline-none focus:border-blue-500"
                placeholder="Enter proposed revision..."
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="text-[11px] text-gray-400">Rationale (Agent Context):</label>
                <VoiceDictationButton
                  onTranscript={(txt) => setRationale((prev) => (prev ? `${prev} ${txt}` : txt))}
                />
              </div>
              <input
                type="text"
                value={rationale}
                onChange={(e) => setRationale(e.target.value)}
                className="w-full bg-[#0b101c] border border-gray-700 rounded px-2.5 py-1.5 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                placeholder="Why should the agent make this change?"
              />
            </div>

            {/* Guardrail Chips */}
            <GuardrailChips
              onSelect={(rule) =>
                setRationale((prev) => (prev ? `${prev}. ${rule}` : rule))
              }
            />

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={resetForms}
                className="px-3 py-1 rounded bg-gray-800 hover:bg-gray-700 text-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={submitReplacement}
                className="px-3 py-1 rounded bg-blue-600 hover:bg-blue-500 text-white font-medium flex items-center gap-1"
              >
                <Check className="w-3.5 h-3.5" />
                Add Suggestion
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Comment Modal */}
      {selection && activeDialog === 'comment' && (
        <div
          className="fixed z-50 transform -translate-x-1/2 bg-[#172033] border border-blue-500/50 rounded-xl shadow-2xl p-4 w-80 text-xs text-slate-200 animate-in fade-in zoom-in-95 duration-150"
          style={{ left: `${selection.x}px`, top: `${Math.max(20, selection.y - 100)}px` }}
        >
          <div className="flex items-center justify-between pb-2 mb-2 border-b border-gray-700">
            <span className="font-semibold text-blue-400 flex items-center gap-1.5">
              <MessageSquare className="w-3.5 h-3.5" />
              Add Comment (Lines {selection.startLine}-{selection.endLine})
            </span>
            <button onClick={resetForms} className="text-gray-400 hover:text-white">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-end">
              <VoiceDictationButton
                onTranscript={(txt) => setCommentText((prev) => (prev ? `${prev} ${txt}` : txt))}
              />
            </div>
            <textarea
              value={commentText}
              onChange={(e) => setCommentText(e.target.value)}
              rows={3}
              autoFocus
              className="w-full bg-[#0b101c] border border-gray-700 rounded p-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
              placeholder="What should the agent know?"
            />
            <div className="flex justify-end gap-2">
              <button
                onClick={resetForms}
                className="px-3 py-1 rounded bg-gray-800 hover:bg-gray-700 text-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={submitComment}
                className="px-3 py-1 rounded bg-blue-600 hover:bg-blue-500 text-white font-medium flex items-center gap-1"
              >
                <Check className="w-3.5 h-3.5" />
                Add Comment
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Annotations Sidebar */}
      <div className="w-84 border-l border-gray-800 bg-[#090e1a] p-4 overflow-y-auto hidden lg:block">
        <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Annotations ({annotations.length})
        </h4>

        {annotations.length === 0 ? (
          <p className="text-xs text-gray-400 italic">No annotations added yet. Select text in the document to begin.</p>
        ) : (
          <div className="space-y-3">
            {annotations.map((ann) => (
              <div
                key={ann.id}
                className="p-3 bg-[#11192e] border border-gray-800 rounded-lg text-xs relative group shadow-sm hover:border-gray-700 transition"
              >
                <button
                  onClick={() => onDeleteAnnotation(ann.id)}
                  className="absolute top-2 right-2 text-gray-400 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition"
                  title="Delete annotation"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>

                <div className="flex items-center gap-1.5 font-medium text-slate-200 mb-1">
                  <span
                    className={`w-2 h-2 rounded-full ${
                      ann.type === 'replacement'
                        ? 'bg-amber-400'
                        : ann.type === 'deletion'
                        ? 'bg-rose-400'
                        : 'bg-blue-400'
                    }`}
                  />
                  <span className="capitalize">{ann.type}</span>
                  <span className="text-gray-400 font-mono text-[11px]">
                    (L{ann.startLine}{ann.endLine !== ann.startLine ? `-${ann.endLine}` : ''})
                  </span>
                </div>

                {ann.type === 'replacement' && (
                  <div className="space-y-1.5 mt-2">
                    {/* Word-level diff visualizer */}
                    <div className="p-2 rounded bg-[#0a0f1d] border border-gray-800 text-[11px] font-mono leading-relaxed space-y-1">
                      <div className="text-rose-300">
                        - {ann.original}
                      </div>
                      <div className="text-emerald-300">
                        + {ann.proposed}
                      </div>
                      <div className="pt-1 border-t border-gray-800/60 text-[10px] text-gray-400">
                        Word diff:{' '}
                        {computeWordDiff(ann.original || '', ann.proposed || '').map((part, pIdx) => (
                          <span
                            key={pIdx}
                            className={
                              part.type === 'added'
                                ? 'bg-emerald-950 text-emerald-300 px-0.5 rounded'
                                : part.type === 'removed'
                                ? 'bg-rose-950 text-rose-300 px-0.5 rounded line-through'
                                : 'text-gray-400'
                            }
                          >
                            {part.value}
                          </span>
                        ))}
                      </div>
                    </div>

                    {ann.rationale && (
                      <p className="text-gray-400 text-[11px] flex items-center gap-1 pt-1">
                        <CornerDownRight className="w-3 h-3 text-gray-400 shrink-0" />
                        <span className="italic">{ann.rationale}</span>
                      </p>
                    )}
                  </div>
                )}

                {ann.type === 'deletion' && (
                  <div className="space-y-1 mt-2">
                    <div className="bg-rose-950/40 text-rose-300 p-1.5 rounded font-mono text-[11px] line-through truncate">
                      {ann.original}
                    </div>
                    {ann.rationale && (
                      <p className="text-gray-400 text-[11px] italic pt-1">{ann.rationale}</p>
                    )}
                  </div>
                )}

                {ann.type === 'comment' && (
                  <p className="text-slate-300 mt-2 bg-[#0c1220] p-2 rounded border border-gray-800">
                    {ann.comment}
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
