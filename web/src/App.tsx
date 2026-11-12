import { useState, useEffect, useCallback } from 'react';
import { Session, Annotation, SessionStatus } from './types';
import { Navbar } from './components/Navbar';
import { SessionTabs } from './components/SessionTabs';
import { PlanAnnotator } from './components/PlanAnnotator';
import { DiffReviewer } from './components/DiffReviewer';
import { ComponentFrame } from './components/ComponentFrame';
import { DiagramPins } from './components/DiagramPins';
import { SurgicalPacketDrawer } from './components/SurgicalPacketDrawer';
import { SubmitModal } from './components/SubmitModal';
import { Loader2 } from 'lucide-react';

export function App() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPacketDrawerOpen, setIsPacketDrawerOpen] = useState(false);
  const [isSubmitModalOpen, setIsSubmitModalOpen] = useState(false);
  const [submitStatus, setSubmitStatus] = useState<SessionStatus>('approved');
  const [submissionComplete, setSubmissionComplete] = useState<string | null>(null);

  // Extract session ID from URL path: /session/:id
  const getSessionIdFromUrl = (): string | null => {
    const match = window.location.pathname.match(/\/session\/([^/]+)/);
    return match ? match[1] : null;
  };

  const getBackendBase = (): string => {
    const params = new URLSearchParams(window.location.search);
    const backendPort = params.get('backend');
    if (backendPort) {
      return `http://127.0.0.1:${backendPort}`;
    }
    return '';
  };

  const fetchSessions = useCallback(async () => {
    try {
      const base = getBackendBase();
      const res = await fetch(`${base}/api/sessions`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: Session[] = await res.json();
      setSessions(data);

      const urlSessionId = getSessionIdFromUrl();
      if (urlSessionId) {
        setActiveSessionId(urlSessionId);
      } else if (data.length > 0 && !activeSessionId) {
        setActiveSessionId(data[0].id);
      }
      setLoading(false);
    } catch (err: any) {
      setError(err?.message || 'Failed to load sessions');
      setLoading(false);
    }
  }, [activeSessionId]);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  // Connect WebSocket for real-time live sync
  useEffect(() => {
    const base = getBackendBase();
    const wsUrl = base
      ? `${base.replace('http', 'ws')}/ws`
      : `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`;

    let socket: WebSocket | null = null;
    try {
      socket = new WebSocket(wsUrl);
      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === 'SESSION_CREATED' || msg.type === 'FEEDBACK_SUBMITTED') {
            fetchSessions();
          }
        } catch {
          // Ignore parse errors
        }
      };
    } catch {
      // WebSocket connection fallback
    }

    return () => {
      if (socket) socket.close();
    };
  }, [fetchSessions]);

  const activeSession = sessions.find((s) => s.id === activeSessionId) || sessions[0] || null;

  const handleAddAnnotation = async (annData: Omit<Annotation, 'id' | 'createdAt'>) => {
    if (!activeSession) return;
    const base = getBackendBase();
    try {
      const res = await fetch(`${base}/api/sessions/${activeSession.id}/annotations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(annData),
      });
      if (res.ok) {
        const created: Annotation = await res.json();
        setSessions((prev) =>
          prev.map((s) =>
            s.id === activeSession.id
              ? { ...s, annotations: [...s.annotations, created] }
              : s
          )
        );
      }
    } catch (e) {
      console.error('Failed to add annotation', e);
    }
  };

  const handleDeleteAnnotation = async (annId: string) => {
    if (!activeSession) return;
    const base = getBackendBase();
    try {
      await fetch(`${base}/api/sessions/${activeSession.id}/annotations/${annId}`, {
        method: 'DELETE',
      });
      setSessions((prev) =>
        prev.map((s) =>
          s.id === activeSession.id
            ? { ...s, annotations: s.annotations.filter((a) => a.id !== annId) }
            : s
        )
      );
    } catch (e) {
      console.error('Failed to delete annotation', e);
    }
  };

  const handleOpenSubmit = (status: SessionStatus) => {
    setSubmitStatus(status);
    setIsSubmitModalOpen(true);
  };

  const handleConfirmSubmit = async (summary: string) => {
    if (!activeSession) return;
    setIsSubmitModalOpen(false);
    const base = getBackendBase();

    try {
      const res = await fetch(`${base}/api/sessions/${activeSession.id}/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          status: submitStatus,
          summary,
          annotations: activeSession.annotations,
        }),
      });

      if (res.ok) {
        setSubmissionComplete(`Feedback successfully dispatched to ${activeSession.agent}.`);
        fetchSessions();
      }
    } catch (e) {
      console.error('Failed to submit feedback', e);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#090d16] flex flex-col items-center justify-center text-slate-300">
        <Loader2 className="w-8 h-8 text-blue-500 animate-spin mb-3" />
        <p className="text-sm font-medium">Initializing PlanMark Studio...</p>
      </div>
    );
  }

  if (error && sessions.length === 0) {
    return (
      <div className="min-h-screen bg-[#090d16] flex flex-col items-center justify-center p-6 text-slate-300">
        <div className="bg-[#111827] border border-gray-800 p-6 rounded-xl max-w-md w-full shadow-2xl text-center">
          <h2 className="text-base font-semibold text-rose-400 mb-2">Connection Error</h2>
          <p className="text-xs text-gray-400 mb-4">{error}</p>
          <button
            onClick={() => fetchSessions()}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-medium transition"
          >
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#090d16] flex flex-col text-slate-100">
      <Navbar
        session={activeSession}
        onOpenPacket={() => setIsPacketDrawerOpen(true)}
        onSubmit={handleOpenSubmit}
        annotationCount={activeSession ? activeSession.annotations.length : 0}
      />

      <SessionTabs
        sessions={sessions}
        activeSessionId={activeSessionId}
        onSelectSession={(id) => {
          setActiveSessionId(id);
          window.history.pushState(null, '', `/session/${id}${window.location.search}`);
        }}
      />

      {submissionComplete && (
        <div className="bg-emerald-950/80 border-b border-emerald-800/80 px-4 py-2.5 flex items-center justify-between text-xs text-emerald-200">
          <span>✓ {submissionComplete}</span>
          <button
            onClick={() => setSubmissionComplete(null)}
            className="text-emerald-400 hover:text-white font-medium"
          >
            Dismiss
          </button>
        </div>
      )}

      <main className="flex-1 overflow-hidden">
        {activeSession ? (
          <>
            {activeSession.artifact.type === 'markdown' && (
              <PlanAnnotator
                content={activeSession.artifact.content}
                annotations={activeSession.annotations}
                onAddAnnotation={handleAddAnnotation}
                onDeleteAnnotation={handleDeleteAnnotation}
              />
            )}

            {activeSession.artifact.type === 'diff' && (
              <DiffReviewer
                rawDiff={activeSession.artifact.content}
                parsedDiff={activeSession.artifact.parsed}
                annotations={activeSession.annotations}
                onAddAnnotation={handleAddAnnotation}
                onDeleteAnnotation={handleDeleteAnnotation}
              />
            )}

            {activeSession.artifact.type === 'html' && (
              <ComponentFrame
                sessionId={activeSession.id}
                annotations={activeSession.annotations}
                onAddAnnotation={handleAddAnnotation}
                onDeleteAnnotation={handleDeleteAnnotation}
              />
            )}

            {activeSession.artifact.type === 'mermaid' && (
              <DiagramPins
                source={activeSession.artifact.content}
                annotations={activeSession.annotations}
                onAddAnnotation={handleAddAnnotation}
                onDeleteAnnotation={handleDeleteAnnotation}
              />
            )}
          </>
        ) : (
          <div className="h-[calc(100vh-3.5rem)] flex flex-col items-center justify-center text-gray-400 p-8 text-center">
            <h3 className="text-base font-semibold text-slate-200 mb-1">No Active Review Sessions</h3>
            <p className="text-xs max-w-sm mb-4">
              Piping a plan, running <code>planmark review &lt;file&gt;</code>, or calling the MCP review tool will populate this studio.
            </p>
          </div>
        )}
      </main>

      <SurgicalPacketDrawer
        isOpen={isPacketDrawerOpen}
        onClose={() => setIsPacketDrawerOpen(false)}
        session={activeSession}
        annotations={activeSession ? activeSession.annotations : []}
      />

      <SubmitModal
        isOpen={isSubmitModalOpen}
        onClose={() => setIsSubmitModalOpen(false)}
        status={submitStatus}
        annotationCount={activeSession ? activeSession.annotations.length : 0}
        onConfirm={handleConfirmSubmit}
      />
    </div>
  );
}
export default App;
