// Minimal DOM helpers for Gitea's MIT-licensed dropdown accessibility patch.
let sequence = 0;
export function generateElemId(prefix: string) {return `${prefix}${++sequence}`;}
export function queryElems(container: ParentNode, selector: string, callback: (element: Element) => void) {container.querySelectorAll(selector).forEach(callback);}
