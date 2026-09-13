// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

func SourceURL(source, id string) string {
	if !ValidID(source, id) {
		return ""
	}
	switch source {
	case "aym":
		encoded := base64.RawURLEncoding.EncodeToString([]byte("kbb:" + strings.SplitN(id, ":", 2)[1]))
		return yargitay.AYMOrigin + "/kbb/pages/search/" + aymType(id) + "?" + url.Values{"id": {encoded}, "type": {aymType(id)}}.Encode()
	case "danistay":
		return yargitay.DanistayOrigin + "/getDokuman?" + url.Values{"id": {id}, "arananKelime": {""}}.Encode()
	case "aihm":
		return yargitay.AIHMOrigin + "/eng?i=" + id
	}
	return ""
}
