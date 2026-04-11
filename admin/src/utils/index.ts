// isEmptyObject
export const isEmptyObject = (obj: Record<string, unknown> | null) => {
  if (!obj) return true;
  return Object.keys(obj).length === 0;
};