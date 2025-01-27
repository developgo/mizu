import styles from "./TimelineBarChart.module.sass";
import { StatsMode } from "../TrafficStatsModal"
import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    Tooltip,
    Legend
} from "recharts";
import { Utils } from "../../../../helpers/Utils";

interface TimelineBarChartProps {
    timeLineBarChartMode: string;
    data: any;
}

export const TimelineBarChart: React.FC<TimelineBarChartProps> = ({ timeLineBarChartMode, data }) => {
    const [protocolStats, setProtocolStats] = useState([]);
    const [protocolsNamesAndColors, setProtocolsNamesAndColors] = useState([]);

    const padTo2Digits = useCallback((num) => {
        return String(num).padStart(2, '0');
    }, [])

    const getHoursAndMinutes = useCallback((protocolTimeKey) => {
        const time = new Date(protocolTimeKey)
        const hoursAndMinutes = padTo2Digits(time.getHours()) + ':' + padTo2Digits(time.getMinutes());
        return hoursAndMinutes;
    }, [padTo2Digits])

    const creatUniqueObjArray = useCallback((objArray) => {
        return [
            ...new Map(objArray.map((item) => [item["name"], item])).values(),
        ];
    }, [])

    useEffect(() => {
        if (!data) return;
        const protocolsBarsData = [];
        const prtcNames = [];
        data.forEach(protocolObj => {
            let obj: { [k: string]: any } = {};
            obj.timestamp = getHoursAndMinutes(protocolObj.timestamp);
            protocolObj.protocols.forEach(protocol => {
                obj[`${protocol.name}`] = protocol[StatsMode[timeLineBarChartMode]];
                prtcNames.push({ name: protocol.name, color: protocol.color });
            })
            protocolsBarsData.push(obj);
        })
        const uniqueObjArray = creatUniqueObjArray(prtcNames);
        protocolsBarsData.sort((a, b) => a.timestamp < b.timestamp ? -1 : 1);
        setProtocolStats(protocolsBarsData);
        setProtocolsNamesAndColors(uniqueObjArray);
    }, [data, timeLineBarChartMode, setProtocolStats, setProtocolsNamesAndColors, creatUniqueObjArray, getHoursAndMinutes])

    const bars = useMemo(() => protocolsNamesAndColors.map((protocolToDIsplay) => {
        return <Bar key={protocolToDIsplay.name} dataKey={protocolToDIsplay.name} stackId="a" fill={protocolToDIsplay.color} />
    }), [protocolsNamesAndColors])

    return (
        <div className={styles.barChartContainer}>
            <BarChart
                width={730}
                height={250}
                data={protocolStats}
                margin={{
                    top: 20,
                    right: 30,
                    left: 20,
                    bottom: 5
                }}
            >
                <XAxis dataKey="timestamp" />
                <YAxis tickFormatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value} />
                <Tooltip formatter={(value) => timeLineBarChartMode === "VOLUME" ? Utils.humanFileSize(value) : value + " Requests"} />
                <Legend />
                {bars}
            </BarChart>
        </div>
    );
}
