import React, { useState } from 'react';
import { Annotation, ParsedDiff, DiffFile } from '../types';
import { Columns, AlignJustify, Plus, Trash2, Check, X, FileCode } from 'lucide-react';

interface DiffReviewerProps {
  parsedDiff?: ParsedDiff;
  rawDiff: string;
  annotations: Annotation[];
  onAddAnnotation: (ann: Omit<Annotation, 'id' | 'createdAt'>) => void;
  onDeleteAnnotation: (id: string) => void;
}

export const DiffReviewer: React.FC<DiffReviewerProps> = ({
  parsedDiff,
  rawDiff,
  annotations,
  onAddAnnotation,
  onDeleteAnnotation,
}) => {
  const [viewMode, setViewMode] = useState<'unified' | 'split'>('unified');
  const [activeFileIndex, setActiveFileIndex] = useState(0);
  const [suggestionLine, setSuggestionLine] = useState<{ file: string; line: number } | null>(null);
  const [suggestionCode, setSuggestionCode] = useState('');
  const [suggestionComment, setSuggestionComment] = useState('');

  // Fallback if parsed diff is not populated
  const files: DiffFile[] = parsedDiff?.files || [
    {
      oldPath: 'diff.patch',
      newPath: 'diff.patch',
      isNew: false,
      isDeleted: false,
      additions: 0,
      deletions: 0,
      hunks: [],
    },
  ];

  const currentFile = files[activeFileIndex] || files[0];

  const handleStartSuggestion = (file: string, line: number, initialContent: string) => {
    setSuggestionLine({ file, line });
    setSuggestionCode(initialContent);
    setSuggestionComment('');
  };

  const submitSuggestion = () => {
    if (!suggestionLine) return;
    onAddAnnotation({
      type: 'suggestion',
      target: suggestionLine.file,
      startLine: suggestionLine.line,
      endLine: suggestionLine.line,
      proposed: suggestionCode,
      comment: suggestionComment || 'Suggested code revision',
    });
    setSuggestionLine(null);
  };

  return (
    <div className="flex h-[calc(100vh-3.5rem)] relative overflow-hidden bg-[#090d16]">
      {/* File Tree Sidebar */}
      <div className="w-64 border-r border-gray-800 bg-[#0c1220] flex flex-col shrink-0">
        <div className="p-3 border-b border-gray-800 flex items-center justify-between text-xs text-gray-400 font-semibold uppercase tracking-wider">
          <span>Files Changed ({files.length})</span>
          {parsedDiff && (
            <span className="font-mono text-[11px]">
              <strong className="text-emerald-400">+{parsedDiff.totalAdditions}</strong>{' '}
              <strong className="text-rose-400">-{parsedDiff.totalDeletions}</strong>
            </span>
          )}
        </div>

        <div className="flex-1 overflow-y-auto divide-y divide-gray-800/40">
          {files.map((file, idx) => {
            const isSelected = idx === activeFileIndex;
            return (
              <button
                key={file.newPath || idx}
                onClick={() => setActiveFileIndex(idx)}
                className={`w-full text-left p-2.5 text-xs flex items-center justify-between transition ${
                  isSelected ? 'bg-gray-800 text-white font-medium' : 'text-gray-400 hover:bg-gray-900 hover:text-gray-200'
                }`}
              >
                <div className="flex items-center gap-2 truncate">
                  <FileCode className="w-3.5 h-3.5 text-blue-400 shrink-0" />
                  <span className="font-mono truncate">{file.newPath || file.oldPath}</span>
                </div>
                <div className="flex items-center gap-1 font-mono text-[10px] shrink-0 ml-2">
                  <span className="text-emerald-400">+{file.additions}</span>
                  <span className="text-rose-400">-{file.deletions}</span>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Main Diff Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Top Control Bar */}
        <div className="h-10 border-b border-gray-800 bg-[#0f172a] px-4 flex items-center justify-between text-xs">
          <div className="flex items-center gap-2 font-mono text-slate-300">
            <span className="font-semibold text-slate-100">{currentFile.newPath || currentFile.oldPath}</span>
            <span className="text-gray-500">|</span>
            <span className="text-emerald-400">+{currentFile.additions}</span>
            <span className="text-rose-400">-{currentFile.deletions}</span>
          </div>

          <div className="flex items-center gap-1 bg-gray-900 p-0.5 rounded border border-gray-800">
            <button
              onClick={() => setViewMode('unified')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded text-xs transition ${
                viewMode === 'unified' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              <AlignJustify className="w-3 h-3" />
              <span>Unified</span>
            </button>
            <button
              onClick={() => setViewMode('split')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded text-xs transition ${
                viewMode === 'split' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              <Columns className="w-3 h-3" />
              <span>Split</span>
            </button>
          </div>
        </div>

        {/* Diff Canvas */}
        <div className="flex-1 overflow-y-auto p-4 font-mono text-xs leading-relaxed bg-[#0a0f1d]">
          {currentFile.hunks.length === 0 ? (
            <div className="bg-[#0f172a] p-4 rounded border border-gray-800 whitespace-pre-wrap text-slate-300">
              {rawDiff}
            </div>
          ) : (
            <div className="space-y-4">
              {currentFile.hunks.map((hunk, hIdx) => (
                <div key={hIdx} className="border border-gray-800 rounded overflow-hidden">
                  <div className="bg-[#131d33] px-3 py-1 text-gray-400 border-b border-gray-800 flex items-center justify-between select-none">
                    <span>{hunk.header}</span>
                  </div>

                  <div className="divide-y divide-gray-800/30 bg-[#0c1222]">
                    {hunk.lines.map((l, lIdx) => {
                      const isAdd = l.type === 'add';
                      const isDel = l.type === 'delete';
                      const targetLine = l.newLine || l.oldLine || 1;
                      const hasSuggestion = suggestionLine?.file === currentFile.newPath && suggestionLine?.line === targetLine;

                      let bg = 'hover:bg-gray-800/50';
                      if (isAdd) bg = 'bg-emerald-950/20 text-emerald-300';
                      if (isDel) bg = 'bg-rose-950/20 text-rose-300';

                      return (
                        <React.Fragment key={lIdx}>
                          <div className={`flex items-center group relative ${bg}`}>
                            {/* Old Line Gutter */}
                            <span className="w-12 text-right pr-2 text-gray-400 select-none py-0.5 border-r border-gray-800/60 shrink-0">
                              {l.oldLine || ''}
                            </span>
                            {/* New Line Gutter */}
                            <span className="w-12 text-right pr-2 text-gray-400 select-none py-0.5 border-r border-gray-800/60 shrink-0">
                              {l.newLine || ''}
                            </span>

                            {/* Suggestion '+' button */}
                            <button
                              onClick={() => handleStartSuggestion(currentFile.newPath, targetLine, l.content)}
                              className="absolute left-20 w-4 h-4 rounded bg-blue-600 text-white flex items-center justify-center opacity-0 group-hover:opacity-100 transition z-10 hover:scale-110 shadow"
                              title="Add code suggestion on this line"
                            >
                              <Plus className="w-3 h-3" />
                            </button>

                            {/* Line prefix */}
                            <span className="w-6 text-center select-none font-bold shrink-0">
                              {isAdd ? '+' : isDel ? '-' : ' '}
                            </span>

                            {/* Line Content */}
                            <span className="flex-1 px-2 whitespace-pre-wrap break-all py-0.5 text-slate-200">
                              {l.content || '\u00A0'}
                            </span>
                          </div>

                          {/* Inline Suggestion Box */}
                          {hasSuggestion && (
                            <div className="bg-[#152038] border-y-2 border-blue-500 p-3 pl-28 text-xs text-slate-100 space-y-2 animate-in fade-in duration-100">
                              <div className="flex items-center justify-between">
                                <span className="font-semibold text-blue-400">
                                  Add Code Suggestion on line {targetLine}
                                </span>
                                <button
                                  onClick={() => setSuggestionLine(null)}
                                  className="text-gray-400 hover:text-white"
                                >
                                  <X className="w-3.5 h-3.5" />
                                </button>
                              </div>

                              <textarea
                                value={suggestionCode}
                                onChange={(e) => setSuggestionCode(e.target.value)}
                                rows={3}
                                className="w-full bg-[#0a0f1d] border border-gray-700 rounded p-2 font-mono text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                                placeholder="Enter suggested replacement code..."
                              />

                              <input
                                type="text"
                                value={suggestionComment}
                                onChange={(e) => setSuggestionComment(e.target.value)}
                                className="w-full bg-[#0a0f1d] border border-gray-700 rounded px-2.5 py-1.5 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
                                placeholder="Add explanation comment (optional)..."
                              />

                              <div className="flex justify-end gap-2 pt-1">
                                <button
                                  onClick={() => setSuggestionLine(null)}
                                  className="px-3 py-1 rounded bg-gray-800 hover:bg-gray-700 text-gray-300"
                                >
                                  Cancel
                                </button>
                                <button
                                  onClick={submitSuggestion}
                                  className="px-3 py-1 rounded bg-blue-600 hover:bg-blue-500 text-white font-medium flex items-center gap-1"
                                >
                                  <Check className="w-3.5 h-3.5" />
                                  Insert Suggestion
                                </button>
                              </div>
                            </div>
                          )}
                        </React.Fragment>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Code Suggestions Sidebar */}
      <div className="w-80 border-l border-gray-800 bg-[#0c1220] p-4 overflow-y-auto hidden xl:block">
        <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Suggestions ({annotations.filter((a) => a.type === 'suggestion').length})
        </h4>

        {annotations.filter((a) => a.type === 'suggestion').length === 0 ? (
          <p className="text-xs text-gray-400 italic">
            Hover over any diff line and click the "+" icon to insert a GitHub-style code suggestion.
          </p>
        ) : (
          <div className="space-y-3">
            {annotations
              .filter((a) => a.type === 'suggestion')
              .map((ann) => (
                <div
                  key={ann.id}
                  className="p-3 bg-[#111a2e] border border-gray-800 rounded-lg text-xs relative group"
                >
                  <button
                    onClick={() => onDeleteAnnotation(ann.id)}
                    className="absolute top-2 right-2 text-gray-400 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>

                  <div className="font-medium text-slate-200 mb-1 flex items-center gap-1.5">
                    <span className="w-2 h-2 rounded-full bg-blue-400" />
                    <span>{ann.target || currentFile.newPath}</span>
                    <span className="text-gray-400 font-mono">:L{ann.startLine}</span>
                  </div>

                  <pre className="bg-[#090e1a] p-2 rounded border border-gray-800 font-mono text-[11px] text-emerald-300 whitespace-pre-wrap my-1.5">
                    {ann.proposed}
                  </pre>

                  {ann.comment && (
                    <p className="text-gray-400 text-[11px] italic mt-1">{ann.comment}</p>
                  )}
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};
