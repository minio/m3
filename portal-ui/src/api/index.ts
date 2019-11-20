// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import request from 'superagent';
import storage from 'local-storage-fallback';

export class API {
    requestWithAuth(method: string, url: string) {
        const token: string = storage.getItem('token')!;
        return request(method, url).set('sessionId', token);
    }

    makeRequest(method: string, url: string, data?: object) {
        return this.requestWithAuth(method, url)
            .send(data)
            .then(res => res.body)
    }

    setAccessToken(token: string): void {
        storage.setItem('token', token);
    }

    login(email: string, password: string, company: string): Promise<string> {
        const url = 'http://localhost:8080/api/v1/users/login';
        return request
            .post(url)
            .send({email: email, password: password, company: company})
            .then((res: any) => {
                if (res.body.jwt_token) {
                    this.setAccessToken(res.body.jwt_token);
                    return res.body.jwt_token;
                } else if (res.body.error) {
                    // throw will be moved to catch block once bad login returns 403
                    throw res.body.error;
                }
            });
    }

    logout() {
        const url = 'http://localhost:8080/api/v1/users/logout';
        return this.makeRequest('POST', url);
    }

    isLoggedIn() {
        return !!storage.getItem('token');
    }
    //signup
}

const api = new API();
export default api;
