//go:build darwin

package macos

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework LocalAuthentication -framework Foundation
#include <LocalAuthentication/LocalAuthentication.h>
#include <CoreFoundation/CoreFoundation.h>

static int touchid_prompt(const char* creason) {
    LAContext *context = [[LAContext alloc] init];
    NSError *error = nil;
    NSString *reason = [NSString stringWithUTF8String:creason];
    if (![context canEvaluatePolicy:LAPolicyDeviceOwnerAuthenticationWithBiometrics error:&error]) {
        // Fallback: user presence (Touch ID, Watch, or password)
        if (![context canEvaluatePolicy:LAPolicyDeviceOwnerAuthentication error:&error]) {
            return -1; // not available
        }
        __block int ok2 = -2;
        [context evaluatePolicy:LAPolicyDeviceOwnerAuthentication
                localizedReason:reason
                          reply:^(BOOL success, NSError * _Nullable err) {
                              ok2 = success ? 1 : 0;
                          }];
        while (ok2 == -2) { [NSThread sleepForTimeInterval:0.05]; }
        return ok2;
    }
    __block int ok = -2;
    [context evaluatePolicy:LAPolicyDeviceOwnerAuthenticationWithBiometrics
            localizedReason:reason
                      reply:^(BOOL success, NSError * _Nullable err) {
                          ok = success ? 1 : 0;
                      }];
    while (ok == -2) { [NSThread sleepForTimeInterval:0.05]; }
    return ok;
}
*/
import "C"

import (
    "errors"
)

func RequireBiometry(reason string) error {
    r := C.touchid_prompt(C.CString(reason))
    if r == 1 { return nil }
    if r == -1 { return errors.New("biometry/user presence not available") }
    return errors.New("biometry/user presence failed")
}