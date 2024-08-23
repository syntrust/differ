// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;

import "./DiffChecker.sol";

contract DiffFix is DiffChecker {
    function expectedDiff(uint256 _interval, uint256 _diff, uint256 _cutoff, uint256 _diffAdjDivisor, uint256 _minDiff)
        internal
        pure
        override
        returns (uint256)
    {
        uint256 diff = _diff;
        if (_interval < _cutoff) {
            diff = diff + (diff - _interval * diff / _cutoff) / _diffAdjDivisor;
            if (diff < _minDiff) {
                diff = _minDiff;
            }
        } else {
            uint256 dec = (_interval * diff / _cutoff - diff) / _diffAdjDivisor;
            if (dec + _minDiff > diff) {
                diff = _minDiff;
            } else {
                diff = diff - dec;
            }
        }
        return diff;
    }
}
