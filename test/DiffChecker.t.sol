// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;

import {Test} from "forge-std/Test.sol";
import {Diff} from "../src/Diff.sol";
import {DiffFix} from "../src/DiffFix.sol";
import {DiffChecker} from "../src/DiffChecker.sol";

contract DiffCheckerTest is Test {
    function test_CheckDiff() public {
        address miner = 0x471977571aD818379E2b6CC37792a5EaC85FdE22;
        uint256 interval = 1724243184 - 1724242128;
        uint256 lastDiff = 71922272;
        uint256 nonce = 870911;
        bytes32 randao = 0xd2c65935c5fc39a515a0282c03ccd8a7879264ae01f697c89da2b0c6e3fa296d;
        bytes32[] memory encodedSamples = new bytes32[](2);
        encodedSamples[0] = 0x0c1e76b42ca04b0b349ed3e6300cc8f67bdfa16f35315acae9d53849935eb727;
        encodedSamples[1] = 0x2a0b69a3663ae7ad5cf2d9da8bdfff961f5631243f5bdc337e1fd20fad5f3e2d;

        DiffChecker diffChecker = new Diff();
        vm.expectRevert("diff not match");
        diffChecker.checkDiff(miner, interval, lastDiff, nonce, randao, encodedSamples);

        diffChecker = new DiffFix();
        diffChecker.checkDiff(miner, interval, lastDiff, nonce, randao, encodedSamples);
    }
}
