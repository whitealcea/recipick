import type { FormTag } from '../types/form';

type TagInputProps = {
  value: FormTag[];
  onChange: (next: FormTag[]) => void;
};

const parseTagLine = (line: string): FormTag | null => {
  const trimmed = line.trim();
  if (!trimmed) return null;

  if (trimmed.startsWith('u:')) {
    const name = trimmed.slice(2).trim();
    return name ? { name, tagType: 'user' } : null;
  }

  if (trimmed.startsWith('s:')) {
    const name = trimmed.slice(2).trim();
    return name ? { name, tagType: 'system' } : null;
  }

  return { name: trimmed, tagType: 'user' };
};

const toText = (tags: FormTag[]): string => {
  return tags
    .map((tag) => `${tag.tagType === 'system' ? 's' : 'u'}:${tag.name}`)
    .join('\n');
};

export const TagInput = ({ value, onChange }: TagInputProps) => {
  return (
    <div className="field">
      <label htmlFor="tags">タグ</label>
      <textarea
        id="tags"
        name="tags"
        rows={5}
        value={toText(value)}
        onChange={(event) => {
          const next = event.target.value
            .split('\n')
            .map(parseTagLine)
            .filter((tag): tag is FormTag => Boolean(tag));
          onChange(next);
        }}
      />
      <p className="hint">1行1タグ。`s:`=system、`u:`=user、接頭辞なしは user。</p>
    </div>
  );
};
