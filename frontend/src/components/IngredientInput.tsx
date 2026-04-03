type IngredientInputProps = {
  value: string;
  onChange: (next: string) => void;
};

export const IngredientInput = ({ value, onChange }: IngredientInputProps) => {
  return (
    <div className="field">
      <label htmlFor="ingredientsText">材料</label>
      <textarea
        id="ingredientsText"
        name="ingredientsText"
        rows={6}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
      <p className="hint">1行1材料で入力します。</p>
    </div>
  );
};
