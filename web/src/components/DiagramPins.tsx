import React, { useEffect, useRef, useState } from 'react';
import mermaid from 'mermaid';
import { Annotation } from '../types';
import { ZoomIn, ZoomOut, RotateCcw, MapPin, Trash2, Check, X } from 'lucide-react';

interface DiagramPinsProps {
  source: string;
  annotations: Annotation[];
  onAddAnnotation: (ann: Omit<Annotation, 'id' | 'createdAt'>) => void;
  onDeleteAnnotation: (id: string) => void;
}

export const DiagramPins: React.FC<DiagramPinsProps> = ({
  source,
  annotations,
  onAddAnnotation,
  onDeleteAnnotation,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [svgContent, setSvgContent] = useState<string>('');
  const [zoom, setZoom] = useState(1);
  const [selectedNode, setSelectedNode] = useState<{ id: string; x: number; y: number } | null>(null);
  const [pinComment, setPinComment] = useState('');
  const [renderError, setRenderError] = useState<string | null>(null);

  useEffect(() => {
    mermaid.initialize({
      startOnLoad: false,
      theme: 'dark',
      securityLevel: 'loose',
      fontFamily: 'Inter, sans-serif',
    });

    const renderDiagram = async () => {
      try {
        setRenderError(null);
        const id = `mermaid_${Date.now()}`;
        const { svg } = await mermaid.render(id, source);
        setSvgContent(svg);
      } catch (err: any) {
        setRenderError(err?.message || 'Failed to render Mermaid diagram');
      }
    };

    renderDiagram();
  }, [source]);

  // Handle diagram click to drop pin
  const handleDiagramClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    // Check if clicked an SVG node element
    let target = e.target as HTMLElement;
    let nodeName = 'diagram';

    const nodeEl = target.closest('.node, .actor, .messageText, [id^="flowchart-"]');
    if (nodeEl) {
      nodeName = nodeEl.id || nodeEl.textContent?.trim().slice(0, 20) || 'node';
    }

    setSelectedNode({
      id: nodeName,
      x,
      y,
    });
  };

  const submitPin = () => {
    if (!selectedNode) return;
    onAddAnnotation({
      type: 'pin',
      target: selectedNode.id,
      comment: pinComment || 'Architecture specification note',
      coordinates: {
        x: selectedNode.x,
        y: selectedNode.y,
        viewport: 1440,
        selector: selectedNode.id,
      },
    });
    setSelectedNode(null);
    setPinComment('');
  };

  return (
    <div className="flex h-[calc(100vh-3.5rem)] relative overflow-hidden bg-[#090d16]">
      {/* Diagram Canvas */}
      <div className="flex-1 flex flex-col items-center">
        {/* Controls */}
        <div className="w-full h-12 border-b border-gray-800 bg-[#0c1220] px-4 flex items-center justify-between">
          <div className="flex items-center gap-2 text-xs text-gray-400">
            <MapPin className="w-3.5 h-3.5 text-amber-400" />
            <span>Click any node or relationship on the diagram to attach visual architectural feedback.</span>
          </div>

          <div className="flex items-center gap-1 bg-gray-900 p-0.5 rounded border border-gray-800">
            <button
              onClick={() => setZoom((z) => Math.max(0.5, z - 0.15))}
              className="p-1.5 text-gray-400 hover:text-white rounded hover:bg-gray-800 transition"
              title="Zoom out"
            >
              <ZoomOut className="w-3.5 h-3.5" />
            </button>
            <span className="text-[11px] font-mono px-2 text-gray-300">
              {Math.round(zoom * 100)}%
            </span>
            <button
              onClick={() => setZoom((z) => Math.min(2.5, z + 0.15))}
              className="p-1.5 text-gray-400 hover:text-white rounded hover:bg-gray-800 transition"
              title="Zoom in"
            >
              <ZoomIn className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setZoom(1)}
              className="p-1.5 text-gray-400 hover:text-white rounded hover:bg-gray-800 transition ml-1"
              title="Reset Zoom"
            >
              <RotateCcw className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Viewport Canvas */}
        <div className="flex-1 w-full p-8 flex items-center justify-center overflow-auto bg-[#070b14] relative">
          {renderError ? (
            <div className="max-w-md p-4 bg-rose-950/60 border border-rose-800 rounded-lg text-rose-300 text-xs font-mono">
              <strong className="block mb-1 text-rose-200">Mermaid Render Error:</strong>
              {renderError}
            </div>
          ) : (
            <div
              ref={containerRef}
              onClick={handleDiagramClick}
              style={{ transform: `scale(${zoom})`, transformOrigin: 'center center' }}
              className="transition-transform duration-100 cursor-crosshair relative max-w-full"
              dangerouslySetInnerHTML={{ __html: svgContent }}
            />
          )}
        </div>
      </div>

      {/* Diagram Pins Sidebar */}
      <div className="w-80 border-l border-gray-800 bg-[#0c1220] p-4 overflow-y-auto flex flex-col">
        <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Diagram Pins ({annotations.filter((a) => a.type === 'pin').length})
        </h4>

        {selectedNode && (
          <div className="mb-4 p-3 bg-[#1e263d] border-2 border-amber-500 rounded-lg text-xs space-y-2 animate-in fade-in duration-150">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-amber-400 flex items-center gap-1">
                <MapPin className="w-3.5 h-3.5" />
                Pin to {selectedNode.id}
              </span>
              <button onClick={() => setSelectedNode(null)} className="text-gray-400 hover:text-white">
                <X className="w-3 h-3" />
              </button>
            </div>

            <textarea
              value={pinComment}
              onChange={(e) => setPinComment(e.target.value)}
              rows={3}
              autoFocus
              className="w-full bg-[#0b101c] border border-gray-700 rounded p-2 text-xs text-slate-100 focus:outline-none focus:border-amber-500"
              placeholder="E.g. This node should be marked as read-replica..."
            />

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setSelectedNode(null)}
                className="px-2.5 py-1 rounded bg-gray-800 hover:bg-gray-700 text-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={submitPin}
                className="px-3 py-1 rounded bg-amber-600 hover:bg-amber-500 text-white font-medium flex items-center gap-1"
              >
                <Check className="w-3.5 h-3.5" />
                Pin Note
              </button>
            </div>
          </div>
        )}

        {annotations.filter((a) => a.type === 'pin').length === 0 ? (
          <p className="text-xs text-gray-400 italic">
            Click on any node, box, or arrow on the diagram to attach visual architectural feedback.
          </p>
        ) : (
          <div className="space-y-3 flex-1 overflow-y-auto">
            {annotations
              .filter((a) => a.type === 'pin')
              .map((ann, i) => (
                <div
                  key={ann.id}
                  className="p-3 bg-[#11192e] border border-gray-800 rounded-lg text-xs relative group shadow-sm"
                >
                  <button
                    onClick={() => onDeleteAnnotation(ann.id)}
                    className="absolute top-2 right-2 text-gray-400 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>

                  <div className="flex items-center gap-2 mb-1">
                    <span className="w-5 h-5 rounded-full bg-amber-600 text-white font-bold flex items-center justify-center text-[10px]">
                      {i + 1}
                    </span>
                    <span className="font-mono text-gray-300 truncate max-w-[170px]">
                      {ann.target}
                    </span>
                  </div>

                  <p className="text-slate-200 mt-2 bg-[#0c1220] p-2 rounded border border-gray-800">
                    {ann.comment}
                  </p>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};
