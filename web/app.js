const out = document.getElementById('out');
function show(v) { out.textContent = typeof v === 'string' ? v : JSON.stringify(v, null, 2); }
async function api(path, opt) {
  const r = await fetch(path, opt);
  const t = await r.text();
  try { return JSON.parse(t); } catch { return t; }
}
document.getElementById('btnRefresh').onclick = async () => {
  show(await api('/api/dashboard'));
};
document.getElementById('btnSync').onclick = async () => {
  show(await api('/api/sync', { method: 'POST' }));
};
document.getElementById('btnSaveRemote').onclick = async () => {
  const url = document.getElementById('remoteUrl').value;
  show(await api('/api/remote', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'origin', url })
  }));
};
api('/api/dashboard').then(show).catch(e => show(String(e)));
