package background

import "fmt"

var ErrBothAccountPOIEmpty = fmt.Errorf("both account number and poi are empty")
var ErrStopRenewWorkflow = fmt.Errorf("workflow does not need to continue")
