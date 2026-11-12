export interface DiffPart {
  type: 'same' | 'added' | 'removed';
  value: string;
}

// Compute word-level differences between original and proposed text
export function computeWordDiff(orig: string, prop: string): DiffPart[] {
  const origWords = orig.split(/(\s+)/);
  const propWords = prop.split(/(\s+)/);

  const parts: DiffPart[] = [];
  let i = 0;
  let j = 0;

  while (i < origWords.length || j < propWords.length) {
    if (i < origWords.length && j < propWords.length && origWords[i] === propWords[j]) {
      parts.push({ type: 'same', value: origWords[i] });
      i++;
      j++;
    } else {
      if (i < origWords.length) {
        parts.push({ type: 'removed', value: origWords[i] });
        i++;
      }
      if (j < propWords.length) {
        parts.push({ type: 'added', value: propWords[j] });
        j++;
      }
    }
  }

  return parts;
}
