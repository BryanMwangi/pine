/**
 * app.js — wires the Todos page to the Pine JSON API.
 *
 * Endpoints used:
 *   GET    /api/todos         → [{id, title, done, created_at}, ...]
 *   POST   /api/todos         → {title: "..."} body → new todo
 *   PATCH  /api/todos/:id     → toggle done
 */

const list     = document.getElementById('todo-list');
const form     = document.getElementById('add-form');
const input    = document.getElementById('new-title');
const emptyMsg = document.getElementById('empty-msg');

// ---- helpers ----

async function apiFetch(path, options = {}) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) throw new Error(`${options.method ?? 'GET'} ${path} → ${res.status}`);
  return res.json();
}

function formatDate(iso) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short', day: 'numeric', year: 'numeric',
  });
}

// ---- render ----

function renderTodo(todo) {
  const li = document.createElement('li');
  li.className = `todo-item${todo.done ? ' done' : ''}`;
  li.dataset.id = todo.id;

  const cb = document.createElement('input');
  cb.type = 'checkbox';
  cb.className = 'todo-checkbox';
  cb.checked = todo.done;
  cb.setAttribute('aria-label', `Mark "${todo.title}" as ${todo.done ? 'incomplete' : 'complete'}`);
  cb.addEventListener('change', () => toggleTodo(todo.id, li));

  const title = document.createElement('span');
  title.className = 'todo-title';
  title.textContent = todo.title;

  const date = document.createElement('span');
  date.className = 'todo-date';
  date.textContent = formatDate(todo.created_at);

  li.append(cb, title, date);
  return li;
}

function syncEmptyState() {
  emptyMsg.hidden = list.children.length > 0;
}

// ---- actions ----

async function loadTodos() {
  const todos = await apiFetch('/api/todos');
  list.innerHTML = '';
  todos.forEach(t => list.appendChild(renderTodo(t)));
  syncEmptyState();
}

async function addTodo(title) {
  const todo = await apiFetch('/api/todos', {
    method: 'POST',
    body: JSON.stringify({ title }),
  });
  list.appendChild(renderTodo(todo));
  syncEmptyState();
}

async function toggleTodo(id, li) {
  const todo = await apiFetch(`/api/todos/${id}`, { method: 'PATCH' });
  li.className = `todo-item${todo.done ? ' done' : ''}`;
  li.querySelector('.todo-checkbox').checked = todo.done;
}

// ---- event wiring ----

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  const title = input.value.trim();
  if (!title) return;
  input.value = '';
  try {
    await addTodo(title);
  } catch (err) {
    console.error('Failed to add todo:', err);
  }
});

// ---- init ----

loadTodos().catch(err => console.error('Failed to load todos:', err));
