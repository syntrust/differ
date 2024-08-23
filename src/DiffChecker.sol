// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;

import {console} from "forge-std/Test.sol";

contract DiffChecker {
    uint256 constant RANDOM_CHECKS = 2;
    uint256 constant CUT_OFF = 7200;
    uint256 constant DIFF_ADJ_DIVISOR = 32;
    uint256 constant MIN_DIFF = 9437184;

    //     uint256 constant SHARD_ENTRY_BITS = 3; // 8 blobs per shard
    //     uint256 constant SAMPLE_SIZE_BITS = 5;
    //     uint256 constant MAX_KV_SIZE_BITS = 17;
    //     uint256 constant SAMPLE_LEN_BITS = MAX_KV_SIZE_BITS - SAMPLE_SIZE_BITS;

    function checkDiff(
        address _miner,
        uint256 _interval,
        uint256 _lastDiff,
        uint256 _nonce,
        bytes32 _randao,
        bytes32[] memory _encodedSamples
    ) public pure {
        bytes32 hash0 = keccak256(abi.encode(_miner, _randao, _nonce));
        for (uint256 i = 0; i < RANDOM_CHECKS; i++) {
            hash0 = keccak256(abi.encode(hash0, _encodedSamples[i]));
        }
        uint256 diff = expectedDiff(_interval, _lastDiff, CUT_OFF, DIFF_ADJ_DIVISOR, MIN_DIFF);
        console.log("diff", diff);
        uint256 required = uint256(2 ** 256 - 1) / diff;
        console.log("required", required);
        console.log("uinthash", uint256(hash0));
        require(uint256(hash0) <= required, "diff not match");
    }

    function expectedDiff(uint256 _interval, uint256 _diff, uint256 _cutoff, uint256 _diffAdjDivisor, uint256 _minDiff)
        internal
        pure
        virtual
        returns (uint256)
    {}
}
