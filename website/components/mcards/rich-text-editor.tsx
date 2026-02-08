'use client';

import { useState } from 'react';

interface RichTextEditorProps {
  content: string;
  onChange: (content: string) => void;
  placeholder?: string;
}

export function RichTextEditor({ content, onChange, placeholder }: RichTextEditorProps) {
  const [text, setText] = useState(content);

  const handleChange = (value: string) => {
    setText(value);
    onChange(JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [{ type: 'text', text: value }],
        },
      ],
    }));
  };

  return (
    <div className="border border-slate-200 dark:border-slate-700 rounded-lg p-4 min-h-[120px]">
      <textarea
        value={text}
        onChange={(e) => handleChange(e.target.value)}
        placeholder={placeholder}
        className="w-full h-full min-h-[100px] resize-none focus:outline-none bg-transparent"
      />
      <p className="text-xs text-slate-400 mt-2">
        TipTap editor will be implemented in a future task
      </p>
    </div>
  );
}
