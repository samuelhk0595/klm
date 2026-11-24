import { request } from '../../engine';

export type GitIdentity = { id: string; alias: string; login: string; isDefault: boolean };
export type GitIdentitiesResponse = { identities: GitIdentity[] };
export type GitIdentityCreate = { alias: string; token: string; makeDefault?: boolean };
export type GitIdentityUpdate = { alias?: string; token?: string; makeDefault?: boolean };

function identitiesPath(projectId: string, identityId?: string) {
  return `/api/projects/${encodeURIComponent(projectId)}/git-identities${identityId === undefined ? '' : `/${encodeURIComponent(identityId)}`}`;
}

export function listGitIdentities(projectId: string, signal?: AbortSignal) {
  return request<GitIdentitiesResponse>(identitiesPath(projectId), 'GET', undefined, 15000, signal);
}

export function testGitIdentity(projectId: string, input: { token: string } | { identityId: string }, signal?: AbortSignal) {
  return request<{ login: string }>(`${identitiesPath(projectId)}/test`, 'POST', input, 15000, signal);
}

export function createGitIdentity(projectId: string, input: GitIdentityCreate, signal?: AbortSignal) {
  return request<GitIdentitiesResponse>(identitiesPath(projectId), 'POST', input, 15000, signal);
}

export function updateGitIdentity(projectId: string, identityId: string, input: GitIdentityUpdate, signal?: AbortSignal) {
  return request<GitIdentitiesResponse>(identitiesPath(projectId, identityId), 'PATCH', input, 15000, signal);
}

export function removeGitIdentity(projectId: string, identityId: string, signal?: AbortSignal) {
  return request<GitIdentitiesResponse>(identitiesPath(projectId, identityId), 'DELETE', undefined, 15000, signal);
}
