export type ArtifactType = 'markdown' | 'diff' | 'html' | 'mermaid';

export type SessionStatus = 'pending' | 'in_review' | 'approved' | 'changes_requested' | 'rejected';

export type AnnotationType = 'replacement' | 'suggestion' | 'deletion' | 'comment' | 'pin';

export interface Point {
  x: number;
  y: number;
  viewport: number;
  selector?: string;
}

export interface Annotation {
  id: string;
  type: AnnotationType;
  target?: string;
  startLine?: number;
  endLine?: number;
  original?: string;
  proposed?: string;
  rationale?: string;
  comment?: string;
  coordinates?: Point;
  createdAt: string;
}

export interface MarkdownLine {
  number: number;
  content: string;
}

export interface MarkdownSection {
  level: number;
  title: string;
  startLine: number;
  endLine: number;
}

export interface ParsedMarkdown {
  lines: MarkdownLine[];
  sections: MarkdownSection[];
  html: string;
}

export interface DiffLine {
  type: 'context' | 'add' | 'delete' | 'header';
  content: string;
  oldLine?: number;
  newLine?: number;
}

export interface DiffHunk {
  header: string;
  oldStart: number;
  oldLines: number;
  newStart: number;
  newLines: number;
  lines: DiffLine[];
}

export interface DiffFile {
  oldPath: string;
  newPath: string;
  isNew: boolean;
  isDeleted: boolean;
  additions: number;
  deletions: number;
  hunks: DiffHunk[];
}

export interface ParsedDiff {
  files: DiffFile[];
  totalAdditions: number;
  totalDeletions: number;
}

export interface MermaidNode {
  id: string;
  label: string;
  line: number;
}

export interface ParsedMermaid {
  type: string;
  nodes: MermaidNode[];
  source: string;
}

export interface Artifact {
  type: ArtifactType;
  title: string;
  filePath?: string;
  content: string;
  parsed?: ParsedMarkdown | ParsedDiff | ParsedMermaid | any;
}

export interface Session {
  id: string;
  agent: string;
  title: string;
  artifact: Artifact;
  status: SessionStatus;
  annotations: Annotation[];
  summary?: string;
  createdAt: string;
  updatedAt: string;
}

export interface FeedbackSubmission {
  status: SessionStatus;
  summary?: string;
  annotations: Annotation[];
}

export interface FeedbackPacket {
  version: string;
  sessionId: string;
  artifact: string;
  status: SessionStatus;
  summary?: string;
  annotations: Annotation[];
  rawMarkdown: string;
}
