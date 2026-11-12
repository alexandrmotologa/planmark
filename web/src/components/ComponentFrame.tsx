import React, { useState, useEffect, useRef } from 'react';
import { Annotation } from '../types';
import { Smartphone, Tablet, Monitor, MousePointer, Trash2, Check, X, MapPin } from 'lucide-react';

interface ComponentFrameProps {
  sessionId: string;
  annotations: Annotation[];
  onAddAnnotation: (ann: Omit<Annotation, 'id' | 'createdAt'>) => void;
  onDeleteAnnotation: (id: string) => void;
}

type ViewportMode = 'mobile' | 'tablet' | 'desktop';

export const ComponentFrame: React.FC<ComponentFrameProps> = ({
  sessionId,
  annotations,
  onAddAnnotation,
  onDeleteAnnotation,
}) => {
  const [viewport, setViewport] = useState<ViewportMode>('desktop');
  const [inspectMode, setInspectMode] = useState(true);
  const [activePin, setActivePin] = useState<{
    selector: string;
    tagName: string;
    x: number;
    y: number;
  } | null>(null);
  const [pinComment, setPinComment] = useState('');
  const iframeRef = useRef<HTMLIFrameElement>(null);

  // Listen to postMessage from iframe bridge script
  useEffect(() => {
    const handleMessage = (e: MessageEvent) => {
      if (e.data && e.data.type === 'PLANMARK_ELEMENT_PINNED') {
        const { selector, tagName, rect } = e.data;
        setActivePin({
          selector: selector || tagName,
          tagName,
          x: rect.x || rect.left,
          y: rect.y || rect.top,
        });
      }
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, []);

  // Update inspect mode in iframe
  useEffect(() => {
    if (iframeRef.current && iframeRef.current.contentWindow) {
      iframeRef.current.contentWindow.postMessage(
        {
          type: 'PLANMARK_SET_INSPECT',
          enabled: inspectMode,
        },
        '*'
      );
    }
  }, [inspectMode]);

  const getViewportWidth = () => {
    switch (viewport) {
      case 'mobile':
        return '390px';
      case 'tablet':
        return '768px';
      case 'desktop':
        return '100%';
    }
  };

  const getViewportNumber = () => {
    switch (viewport) {
      case 'mobile':
        return 390;
      case 'tablet':
        return 768;
      case 'desktop':
        return 1440;
    }
  };

  const submitPin = () => {
    if (!activePin) return;
    onAddAnnotation({
      type: 'pin',
      target: activePin.selector,
      comment: pinComment || `Issue with <${activePin.tagName}> element`,
      coordinates: {
        x: activePin.x,
        y: activePin.y,
        viewport: getViewportNumber(),
        selector: activePin.selector,
      },
    });
    setActivePin(null);
    setPinComment('');
  };

  return (
    <div className="flex h-[calc(100vh-3.5rem)] relative overflow-hidden bg-[#090d16]">
      <div className="flex-1 flex flex-col items-center">
        {/* Viewport Control Bar */}
        <div className="w-full h-12 border-b border-gray-800 bg-[#0c1220] px-4 flex items-center justify-between">
          <div className="flex items-center gap-1 bg-gray-900 p-0.5 rounded border border-gray-800">
            <button
              onClick={() => setViewport('mobile')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded text-xs transition ${
                viewport === 'mobile' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              <Smartphone className="w-3.5 h-3.5" />
              <span>Mobile (390px)</span>
            </button>
            <button
              onClick={() => setViewport('tablet')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded text-xs transition ${
                viewport === 'tablet' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              <Tablet className="w-3.5 h-3.5" />
              <span>Tablet (768px)</span>
            </button>
            <button
              onClick={() => setViewport('desktop')}
              className={`flex items-center gap-1 px-2.5 py-1 rounded text-xs transition ${
                viewport === 'desktop' ? 'bg-blue-600 text-white font-medium' : 'text-gray-400 hover:text-white'
              }`}
            >
              <Monitor className="w-3.5 h-3.5" />
              <span>Desktop (1440px)</span>
            </button>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setInspectMode(!inspectMode)}
              className={`flex items-center gap-1.5 px-3 py-1 rounded text-xs font-medium transition border ${
                inspectMode
                  ? 'bg-blue-950/80 text-blue-300 border-blue-600/80'
                  : 'bg-gray-800 text-gray-400 border-gray-700 hover:text-white'
              }`}
            >
              <MousePointer className="w-3.5 h-3.5" />
              <span>{inspectMode ? 'Inspect Mode ON' : 'Inspect Mode OFF'}</span>
            </button>
          </div>
        </div>

        {/* Viewport Frame Container */}
        <div className="flex-1 w-full p-4 flex items-center justify-center overflow-auto bg-[#070b14]">
          <div
            className="h-full bg-slate-900 rounded-lg shadow-2xl border border-gray-800 overflow-hidden relative transition-all duration-300"
            style={{ width: getViewportWidth() }}
          >
            <iframe
              ref={iframeRef}
              src={`/api/sessions/${sessionId}/sandbox`}
              title="Component Preview"
              className="w-full h-full border-0"
              sandbox="allow-scripts allow-same-origin"
            />
          </div>
        </div>
      </div>

      {/* Visual Pins Sidebar */}
      <div className="w-80 border-l border-gray-800 bg-[#0c1220] p-4 overflow-y-auto flex flex-col">
        <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
          Visual Pins ({annotations.filter((a) => a.type === 'pin').length})
        </h4>

        {activePin && (
          <div className="mb-4 p-3 bg-[#17223b] border-2 border-blue-500 rounded-lg text-xs space-y-2 animate-in fade-in duration-150">
            <div className="flex items-center justify-between">
              <span className="font-semibold text-blue-400 flex items-center gap-1">
                <MapPin className="w-3.5 h-3.5" />
                Add Pin to {activePin.tagName}
              </span>
              <button onClick={() => setActivePin(null)} className="text-gray-400 hover:text-white">
                <X className="w-3 h-3" />
              </button>
            </div>

            <div className="font-mono text-[10px] text-gray-300 bg-gray-900 p-1.5 rounded truncate">
              {activePin.selector}
            </div>

            <textarea
              value={pinComment}
              onChange={(e) => setPinComment(e.target.value)}
              rows={3}
              autoFocus
              className="w-full bg-[#0b101c] border border-gray-700 rounded p-2 text-xs text-slate-100 focus:outline-none focus:border-blue-500"
              placeholder="What should be modified on this UI element?"
            />

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setActivePin(null)}
                className="px-2.5 py-1 rounded bg-gray-800 hover:bg-gray-700 text-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={submitPin}
                className="px-3 py-1 rounded bg-blue-600 hover:bg-blue-500 text-white font-medium flex items-center gap-1"
              >
                <Check className="w-3.5 h-3.5" />
                Save Pin
              </button>
            </div>
          </div>
        )}

        {annotations.filter((a) => a.type === 'pin').length === 0 ? (
          <p className="text-xs text-gray-400 italic">
            Click on any rendered button, card, or element in the preview frame to drop a visual feedback pin.
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
                    <span className="w-5 h-5 rounded-full bg-blue-600 text-white font-bold flex items-center justify-center text-[10px]">
                      {i + 1}
                    </span>
                    <span className="font-mono text-gray-300 truncate max-w-[170px]">
                      {ann.coordinates?.selector || ann.target}
                    </span>
                  </div>

                  <p className="text-slate-200 mt-2 bg-[#0c1220] p-2 rounded border border-gray-800">
                    {ann.comment}
                  </p>

                  {ann.coordinates && (
                    <div className="mt-1 text-[10px] text-gray-400 font-mono">
                      Viewport: {ann.coordinates.viewport}px | X: {ann.coordinates.x.toFixed(0)}, Y:{' '}
                      {ann.coordinates.y.toFixed(0)}
                    </div>
                  )}
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};
