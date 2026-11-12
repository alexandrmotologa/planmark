import React from 'react';

interface GuardrailChipsProps {
  onSelect: (rule: string) => void;
}

export const GuardrailChips: React.FC<GuardrailChipsProps> = ({ onSelect }) => {
  const guardrails = [
    {
      label: '⚡ Single-Binary',
      text: 'Keep implementation single-binary. Do not introduce external daemon or database dependencies.',
    },
    {
      label: '🛡️ Context & Timeouts',
      text: 'Propagate context.Context across all I/O calls and enforce sensible timeouts.',
    },
    {
      label: '🧪 TDD First',
      text: 'Write deterministic unit tests verifying this interface contract before finalizing implementation.',
    },
    {
      label: '📦 Lean Dependencies',
      text: 'Avoid adding new heavy third-party packages; prefer standard library or minimal utilities.',
    },
    {
      label: '🔒 Input Validation',
      text: 'Sanitize parameters, check for nil/empty values, and return typed errors.',
    },
  ];

  return (
    <div className="space-y-1">
      <span className="text-[10px] text-gray-400 font-medium uppercase tracking-wider block">
        Quick Guardrail Presets:
      </span>
      <div className="flex flex-wrap gap-1.5">
        {guardrails.map((g) => (
          <button
            key={g.label}
            type="button"
            onClick={() => onSelect(g.text)}
            className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-[#131d33] hover:bg-blue-900/60 text-blue-300 hover:text-white border border-blue-900/50 hover:border-blue-600 transition"
            title={g.text}
          >
            {g.label}
          </button>
        ))}
      </div>
    </div>
  );
};
